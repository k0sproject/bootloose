# SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
# Test that common utilities are present in the base images

bootloose config create --config %testName.bootloose --override --name %testName --key %testName-key --image %image --volume /lib/modules:/lib/modules:ro --privileged
%defer rm -f %testName.bootloose %testName-key %testName-key.pub
%defer bootloose delete --config %testName.bootloose
bootloose create --config %testName.bootloose

%1 bootloose --config %testName.bootloose ssh root@node0 hostname
%assert equal %1 node0

# test uname and capture buffer interpolation
bootloose --config %testName.bootloose ssh root@%1 -- uname -a

bootloose --config %testName.bootloose ssh root@node0 ps
bootloose --config %testName.bootloose ssh root@node0 ifconfig
bootloose --config %testName.bootloose ssh root@node0 ip route
bootloose --config %testName.bootloose ssh root@node0 -- netstat -n -l
bootloose --config %testName.bootloose ssh root@node0 -- command -v ping
bootloose --config %testName.bootloose ssh root@node0 -- curl --version
bootloose --config %testName.bootloose ssh root@node0 -- command -v wget
bootloose --config %testName.bootloose ssh root@node0 -- vi --help
bootloose --config %testName.bootloose ssh root@node0 -- sudo -n true || doas -n true
