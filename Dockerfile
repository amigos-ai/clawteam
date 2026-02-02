FROM openclaw:local

# Git + SSH essentials
RUN apt-get update && apt-get install -y \
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

ENTRYPOINT ["/entrypoint.sh"]
