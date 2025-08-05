# SPDX-FileCopyrightText: 2025 bootloose authors
# SPDX-License-Identifier: Apache-2.0
FROM fedora:42

ENV container docker

RUN dnf -y install sudo openssh-server procps-ng hostname net-tools iproute iputils wget && dnf clean all

EXPOSE 22

RUN echo -e '#!/bin/sh\n[ -f "/etc/machine-id" ] && rm -f /etc/machine-id\nsystemd-machine-id-setup &> /dev/null\nexec "$@"' > /entry.sh \
    && chmod +x /entry.sh \
    && truncate -c -s0 /etc/issue /etc/motd \

STOPSIGNAL SIGRTMIN+3

ENTRYPOINT ["/entry.sh"]

CMD ["/bin/bash"]
