FROM ghcr.io/openclaw/openclaw:2026.1.30

USER root

# Git + SSH essentials (Debian-based image)
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    openssh-client \
    && rm -rf /var/lib/apt/lists/*

# SSH config for mounted keys
RUN mkdir -p /home/node/.ssh && \
    chmod 700 /home/node/.ssh && \
    printf "Host *\n  IdentityFile /home/node/.ssh/id_rsa\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null\n" \
    > /home/node/.ssh/config && \
    chown -R node:node /home/node/.ssh

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

USER node

ENTRYPOINT ["/entrypoint.sh"]
CMD ["node", "openclaw.mjs", "gateway", "--allow-unconfigured", "--bind", "lan"]
