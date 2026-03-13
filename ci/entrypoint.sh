#!/bin/bash
set -e
if [ -n "${GPG_SIGNING_KEY}" ]; then
  echo "${GPG_SIGNING_KEY}" > /tmp/gpg.key
  chmod 600 /tmp/gpg.key
  export GPG_KEY_PATH=/tmp/gpg.key
fi
exec "$@"
