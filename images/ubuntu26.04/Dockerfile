# SPDX-FileCopyrightText: 2026 bootloose authors
# SPDX-License-Identifier: Apache-2.0
FROM ubuntu:26.04

ENV container=docker

# Don't start any optional services except for the few we need.
RUN set -eux \
  && export DEBIAN_FRONTEND=noninteractive \
  && find /etc/systemd/system \
    /lib/systemd/system \
    -path '*.wants/*' \
    -not -name '*journald*' \
    -not -name '*systemd-tmpfiles*' \
    -not -name '*systemd-user-sessions*' \
    -printf '* Deleting %P\n' -delete \
  && apt-get update \
  && apt-get install -y \
    curl \
    dbus \
    iproute2 \
    iputils-ping \
    net-tools \
    openssh-server \
    sudo \
    systemd \
    vim-tiny \
    wget \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/* \
  && truncate -s0 /etc/machine-id \
  && truncate -s0 /var/lib/dbus/machine-id \
  && systemctl set-default multi-user.target \
  && systemctl mask \
    dev-hugepages.mount \
    sys-fs-fuse-connections.mount \
    systemd-update-utmp.service \
    systemd-tmpfiles-setup.service \
    console-getty.service \
  && systemctl enable ssh.service

# This container image doesn't have locales installed. Disable forwarding the
# user locale env variables or we get warnings such as: bash: warning:
# setlocale: LC_ALL: cannot change locale
RUN sed -i -e 's/^AcceptEnv LANG LC_\*$/#AcceptEnv LANG LC_*/' /etc/ssh/sshd_config

EXPOSE 22

# https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
STOPSIGNAL SIGRTMIN+3

CMD ["/bin/bash"]
