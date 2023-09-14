FROM docker.io/library/alpine:3.18.3

RUN apk add --no-cache \
  alpine-base \
  openssh-server \
  bash \
  shadow \
  doas \
  && rc-update add syslog boot \
  && rc-update add sshd default \
  && rc-update add local default \
# disable ttys
  && sed -i -e 's/^\(tty[0-9]\)/# \1/' /etc/inittab \
# prevent start-stop-daemon from hanging when max_fds is huge
  && sed -Ei -e 's/^[# ](rc_ulimit)=.*/\1="-n 16384"/' /etc/rc.conf \
# no greetings
  && truncate -c -s0 /etc/issue /etc/motd
