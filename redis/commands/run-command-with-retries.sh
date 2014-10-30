#!/bin/bash

#
# Script to install run-command-with-retries.sh to the install directory.
#
# Expected environment variable:
#   DEPLOY_INSTALL_TEMP: the absolute path where run-command-with-retries.sh is installed.

DEPLOY_INSTALL_TEMP=${DEPLOY_INSTALL_TEMP:-/tmp}
mkdir -p ${DEPLOY_INSTALL_TEMP}

cat > ${DEPLOY_INSTALL_TEMP}/run-command-with-retries.sh << 'EOF'
#!/bin/bash

MAX_RETRY_SEC=${MAX_RETRY_SEC:-120}
SLEEP_SEC=${SLEEP_SEC:-2}

# Expect
# COMMAND
readonly COMMAND=$1

# Exit if no command exists
if [ -z "${COMMAND}" ]; then
  exit 1
fi

echo "Running command: '${COMMAND}'"

# Run the command
for ((i = 0; i < ${MAX_RETRY_SEC}; i += ${SLEEP_SEC})); do
  if ${COMMAND}; then
    break
  fi

  echo "Command '${COMMAND}' failed -  retry after sleep ${SLEEP_SEC} seconds..."
  sleep ${SLEEP_SEC}
done
EOF

chmod +x ${DEPLOY_INSTALL_TEMP}/run-command-with-retries.sh
