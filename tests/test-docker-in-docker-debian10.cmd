# SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
%defer bootloose delete --config %testName.yaml
bootloose create --config %testName.yaml
bootloose --config %testName.yaml ssh root@node0 -- apt update && apt install -y docker.io
bootloose --config %testName.yaml ssh root@node0 systemctl start docker
bootloose --config %testName.yaml ssh root@node0 docker pull busybox
%out bootloose --config %testName.yaml ssh root@node0 docker run busybox echo success
