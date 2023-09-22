FROM docker.io/library/alpine:3.18.3

RUN apk add --no-cache \
  alpine-base \
  dbus \
  openssh-server \
  bash \
  shadow \
  doas \
  curl \
  && rc-update add syslog boot \
  && rc-update add sshd default \
  && rc-update add local default \
# disable ttys
  && sed -i -e 's/^\(tty[0-9]\)/# \1/' /etc/inittab \
# prevent start-stop-daemon from hanging when max_fds is huge
  && sed -Ei -e 's/^[# ](rc_ulimit)=.*/\1="-n 16384"/' /etc/rc.conf \
# no greetings
  && truncate -c -s0 /etc/issue /etc/motd \
# allow root to use doas
  && echo permit :wheel >> /etc/doas.d/doas.conf \
  && echo permit nopass keepenv root >> /etc/doas.d/doas.conf

# This container image doesn't have locales installed. Disable forwarding the
# user locale env variables or we get warnings such as:
#  bash: warning: setlocale: LC_ALL: cannot change locale
RUN sed -i -e 's/^AcceptEnv LANG LC_\*$/#AcceptEnv LANG LC_*/' /etc/ssh/sshd_config

# Add entry script
RUN rm -f /etc/machine-id; echo -e '#!/bin/sh\n[ -f "/var/lib/dbus/machine-id" ] && rm /var/lib/dbus/machine-id\ndbus-uuidgen --ensure=/var/lib/dbus/machine-id\nexec "$@"' > /entry.sh \
    && chmod +x /entry.sh

# https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
STOPSIGNAL SIGRTMIN+3

# Set the entry script as the entry point
ENTRYPOINT ["/entry.sh"]

CMD ["/bin/bash"]
