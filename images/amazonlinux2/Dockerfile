FROM amazonlinux:2

ENV container docker

RUN yum -y install sudo systemd hostname procps-ng net-tools iproute iputils wget dbus && yum clean all

RUN (cd /lib/systemd/system/sysinit.target.wants/; for i in *; do [ $i == \
systemd-tmpfiles-setup.service ] || rm -f $i; done); \
rm -f /lib/systemd/system/multi-user.target.wants/*;\
rm -f /etc/systemd/system/*.wants/*;\
rm -f /lib/systemd/system/local-fs.target.wants/*; \
rm -f /lib/systemd/system/sockets.target.wants/*udev*; \
rm -f /lib/systemd/system/sockets.target.wants/*initctl*; \
rm -f /lib/systemd/system/basic.target.wants/*;\
rm -f /lib/systemd/system/anaconda.target.wants/*;\
rm -f /lib/systemd/system/*.wants/*update-utmp*;

RUN yum -y install openssh-server && yum clean all

RUN echo -e '#!/bin/sh\n[ -f "/etc/machine-id" ] && rm -f /etc/machine-id\ndbus-uuidgen --ensure=/etc/machine-id\nexec "$@"' > /entry.sh \
    && chmod +x /entry.sh

EXPOSE 22

# https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
STOPSIGNAL SIGRTMIN+3

# Set the entry script as the entry point
ENTRYPOINT ["/entry.sh"]

CMD ["/bin/bash"]
