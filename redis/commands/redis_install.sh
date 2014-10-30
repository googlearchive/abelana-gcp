#!/bin/bash
# Copyright 2014 Google Inc. All rights reserved.
#
# redis_install.sh
#
# This script installs Redis.

set -o nounset
set -o errexit

# Create the temporary install directory
mkdir -p ${DEPLOY_INSTALL_TEMP}

# For now, we'll focus on just Redis 2.8. We could later provide Redis
# Cluster or some older version.
readonly REDIS_SRC=${REDIS_2_8_SRC}
readonly REDIS_DEB=${REDIS_2_8_DEB}
readonly REDIS_TOOLS_DEB=${REDIS_TOOLS_2_8_DEB}

# Download the Redis Source
${DEPLOY_INSTALL_TEMP}/run-command-with-retries.sh \
  "wget ${REPOSITORY}/${REDIS_SRC} --directory-prefix=${DEPLOY_INSTALL_TEMP}"

# Pull down the debian packages for Jemalloc and Redis
${DEPLOY_INSTALL_TEMP}/run-command-with-retries.sh \
  "wget ${REPOSITORY}/${JEMALLOC_DEB} --directory-prefix=${DEPLOY_INSTALL_TEMP}"

${DEPLOY_INSTALL_TEMP}/run-command-with-retries.sh \
  "wget ${REPOSITORY}/${JEMALLOC_SRC} --directory-prefix=${DEPLOY_INSTALL_TEMP}"

${DEPLOY_INSTALL_TEMP}/run-command-with-retries.sh \
  "wget ${REPOSITORY}/${REDIS_SRC} --directory-prefix=${DEPLOY_INSTALL_TEMP}"

${DEPLOY_INSTALL_TEMP}/run-command-with-retries.sh \
  "wget ${REPOSITORY}/${REDIS_DEB} --directory-prefix=${DEPLOY_INSTALL_TEMP}"

${DEPLOY_INSTALL_TEMP}/run-command-with-retries.sh \
  "wget ${REPOSITORY}/${REDIS_TOOLS_DEB} --directory-prefix=${DEPLOY_INSTALL_TEMP}"

# Avoid the Redis server starting when installing it.
echo exit 101 > /usr/sbin/policy-rc.d
chmod 755 /usr/sbin/policy-rc.d

# Install the software
dpkg --install ${DEPLOY_INSTALL_TEMP}/${JEMALLOC_DEB}
dpkg --install ${DEPLOY_INSTALL_TEMP}/${REDIS_DEB} ${DEPLOY_INSTALL_TEMP}/${REDIS_TOOLS_DEB}

# Remove policy-rc.d. We dont need it anymore.
rm -f /usr/sbin/policy-rc.d

tar -xzf ${DEPLOY_INSTALL_TEMP}/${JEMALLOC_SRC} -C /usr/src
tar -xzf ${DEPLOY_INSTALL_TEMP}/${REDIS_SRC} -C /usr/src
