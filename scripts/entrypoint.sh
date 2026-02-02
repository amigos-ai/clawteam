#!/bin/bash
set -e

# Link SSH key if provided
if [ -f /secrets/ssh/id_rsa ]; then
    cp /secrets/ssh/id_rsa /home/node/.ssh/id_rsa
    chmod 600 /home/node/.ssh/id_rsa
    chown node:node /home/node/.ssh/id_rsa
fi

# Configure git if details provided
if [ -n "$GIT_USER_NAME" ]; then
    git config --file /home/node/.gitconfig user.name "$GIT_USER_NAME"
fi
if [ -n "$GIT_USER_EMAIL" ]; then
    git config --file /home/node/.gitconfig user.email "$GIT_USER_EMAIL"
fi
# Ensure node owns the gitconfig
if [ -f /home/node/.gitconfig ]; then
    chown node:node /home/node/.gitconfig
fi

# Start OpenClaw (existing entrypoint)
exec docker-entrypoint.sh "$@"
