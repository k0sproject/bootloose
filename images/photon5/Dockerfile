FROM photon:5.0

ENV container docker

RUN yum -y install sudo systemd dbus procps-ng net-tools iproute iputils wget vim openssh-server coreutils gawk && yum clean all

# Don't start any optional services except for the few we need.
RUN find /etc/systemd/system \
    /lib/systemd/system \
    -path '*.wants/*' \
    -not -name '*journald*' \
    -not -name '*systemd-tmpfiles*' \
    -not -name '*systemd-user-sessions*' \
    -exec rm \{} \;

RUN echo 'root:*:17995::::::' > /etc/shadow

# Truncate machine ID files to trigger regeneration.
RUN >/etc/machine-id

EXPOSE 22

RUN ssh-keygen -f /etc/ssh/ssh_host_rsa_key -N '' -q -t rsa; \
    ssh-keygen -f /etc/ssh/ssh_host_ecdsa_key -N '' -q -t ecdsa; \
    ssh-keygen -f /etc/ssh/ssh_host_ed25519_key -N '' -q -t ed25519; \
    sed -i 's/^PermitRootLogin no/PermitRootLogin yes/' /etc/ssh/sshd_config; \
    systemctl enable sshd

RUN systemctl set-default multi-user.target; \
    systemctl mask \
    dev-hugepages.mount \
    sys-fs-fuse-connections.mount \
    systemd-update-utmp.service \
    systemd-tmpfiles-setup.service \
    console-getty.service

# https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
STOPSIGNAL SIGRTMIN+3

CMD ["/bin/bash"]

