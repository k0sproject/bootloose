FROM fedora:38

ENV container docker

RUN dnf -y install sudo openssh-server procps-ng hostname net-tools iproute iputils wget && dnf clean all

EXPOSE 22

RUN echo -e '#!/bin/sh\n[ -f "/etc/machine-id" ] && rm -f /etc/machine-id\nsystemd-machine-id-setup &> /dev/null\nexec "$@"' > /entry.sh \
    && chmod +x /entry.sh

# https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
STOPSIGNAL SIGRTMIN+3

# Set the entry script as the entry point
ENTRYPOINT ["/entry.sh"]

CMD ["/bin/bash"]
