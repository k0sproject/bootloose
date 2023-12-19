# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
FROM alpine:3.18.5

RUN apk add --no-cache \
  alpine-base \
  openssh-server \
  doas \
  curl \
  && rc-update add syslog boot \
  && rc-update add machine-id boot \
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

# Start a shell if no command is given, otherwise exec the command verbatim.
ENTRYPOINT ["/bin/sh", "-c", "[ $# -gt 0 ] || set -- /bin/sh; exec \"$@\"", "--"]
