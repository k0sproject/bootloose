# SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
FROM clearlinux:latest

ENV container docker

RUN swupd bundle-add openssh-server vim network-basic sudo
RUN echo 'root:*:17995::::::' > /etc/shadow

EXPOSE 22

RUN echo -e '#!/bin/sh\n[ -f "/etc/machine-id" ] && rm -f /etc/machine-id\ndbus-uuidgen --ensure=/etc/machine-id\nexec "$@"' > /entry.sh \
    && chmod +x /entry.sh

STOPSIGNAL SIGRTMIN+3

# Set the entry script as the entry point
ENTRYPOINT ["/entry.sh"]

CMD ["/bin/bash"]
