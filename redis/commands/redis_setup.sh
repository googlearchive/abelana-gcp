#!/bin/bash
# Copyright 2014 Google Inc. All rights reserved.
#
# redis_install.sh
#
# This script sets up Redis. It performs the following actions:
# * Creates common configuration files for Redis nodes.
# * Formats and mounts disks.
# * Creates specific configurations for slave and persistent Redis nodes.
# * Configures Redis Sentinel.
# * Starts Redis server and sentinel services.

source $GCLOUD_ROOT/pathfile

set -o nounset
set -o errexit

source ${DEPLOY_INSTALL_TEMP}/deployment_util.sh

# Turn on memory overcommit to avoid failing fork on saving db to disk.
# http://redis.io/topics/faq
echo 1 > /proc/sys/vm/overcommit_memory

# Overwrite main config file. This main config could use input from
# some parameters in the deploy screen, such as AOF or RDB formats.
readonly NODE_CONFIG=/etc/redis/redis_node.conf

readonly REDIS_PORT=${REDIS_PORT:-6379}
echo "
port $REDIS_PORT
daemonize yes
pidfile /var/run/redis/redis-server.pid
tcp-backlog 511
timeout 0
tcp-keepalive 0
loglevel notice
logfile /var/log/redis/redis-server.log
databases 16
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dir /var/lib/redis
slave-serve-stale-data yes
slave-read-only yes
appendonly no
dbfilename dump.rdb

include $NODE_CONFIG
" > /etc/redis/redis.conf

touch $NODE_CONFIG

# This is the code that sets up masterships, slaves and db-savers
# First instance is always the master. Last 2 instances should be the slaves. Minimal configuration
# for master/slave is 3 nodes, so in this case, both slaves will persist data to disk.
readonly INSTANCES=$(get_instance_list \
  "${REDIS_NODE_VIEW_NAME}" \
  "${REDIS_ZONE}" | tr '\n' ' ')

if [[ $? != 0 ]]; then
  echo "Failed to get list of instances for ${REDIS_NODE_VIEW_NAME}"
  exit 2
fi

echo "Redis instances: ${INSTANCES}"

readonly NODES=(${INSTANCES})
readonly MASTER=${NODES[0]}
readonly LEN=${#NODES[@]}
readonly SENTINEL_QUORUM=$(((LEN / 2) + 1))

# Figure out number of slaves and persistent nodes.
PERSISTENT_NODES=${REDIS_PERSISTENT_NODES}
declare -a WRITE_NODES

# Make sure we don't exceed the number of nodes.
if [[ ${PERSISTENT_NODES} > ${LEN} ]] || [[ ${PERSISTENT_NODES} == "-1" ]]; then
  PERSISTENT_NODES=${#NODES[@]}
fi

# Only set up write nodes if number of persistent nodes is gt 0.
if [[ ${PERSISTENT_NODES} > 0 ]]; then
  for ((i = 1; i < ${PERSISTENT_NODES} + 1; i++)); do
    WRITE_NODES[i]=${NODES[LEN - i]}
  done

  # Set up write nodes
  for write_node in ${WRITE_NODES[@]}; do
    if [[ $(hostname) == ${write_node} ]]; then
      # These are the default values.
      # We could provide values passed from the deploy iface.
      echo "save 300 10" >> ${NODE_CONFIG}
      echo "save 60 10000" >> ${NODE_CONFIG}
    fi
  done
fi

# Set up slave nodes
if [[ $(hostname) != ${MASTER} ]]; then
  echo "slaveof ${MASTER} ${REDIS_PORT}" >> $NODE_CONFIG
fi

# Set up Sentinel init script
readonly RC_SENTINEL=/etc/init.d/redis-sentinel
cp /etc/init.d/redis-server ${RC_SENTINEL}
sed 's:redis-server:redis-sentinel:g' -i ${RC_SENTINEL}
sed 's:redis.conf:sentinel.conf:g' -i ${RC_SENTINEL}

cat > /etc/redis/sentinel.conf <<EOF
daemonize yes
sentinel monitor master $MASTER 6379 $SENTINEL_QUORUM
sentinel down-after-milliseconds master 60000
sentinel failover-timeout master 180000
sentinel parallel-syncs master 1
EOF

chown redis:redis /etc/redis/sentinel.conf

/etc/init.d/redis-server start
/etc/init.d/redis-sentinel start
