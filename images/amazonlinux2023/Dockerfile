# SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
FROM amazonlinux:2023

ENV container docker

RUN yum -y install sudo systemd hostname procps-ng net-tools iproute iputils wget dbus vim-minimal && yum clean all

RUN (cd /lib/systemd/system/sysinit.target.wants/; for i in *; do [ $i == \
systemd-tmpfiles-setup.service ] || rm -f $i; done); \
rm -f /lib/systemd/system/multi-user.target.wants/*;\
rm -f /etc/systemd/system/*.wants/*;\
rm -f /lib/systemd/system/local-fs.target.wants/*; \
rm -f /lib/systemd/system/sockets.target.wants/*udev*; \
rm -f /lib/systemd/system/sockets.target.wants/*initctl*; \
rm -f /lib/systemd/system/basic.target.wants/*;\
rm -f /lib/systemd/system/anaconda.target.wants/*;\
rm -f /lib/systemd/system/*.wants/*update-utmp*; \
rm -f /var/lib/dbus/machine-id /etc/machine-id

RUN yum -y install openssh-server && yum clean all

RUN echo -e '#!/bin/sh\n[ -f "/etc/machine-id" ] && rm -f /etc/machine-id\nsystemd-machine-id-setup &> /dev/null\nexec "$@"' > /entry.sh \
    && chmod +x /entry.sh

EXPOSE 22

# https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
STOPSIGNAL SIGRTMIN+3

# Set the entry script as the entry point
ENTRYPOINT ["/entry.sh"]

CMD ["/bin/bash"]
