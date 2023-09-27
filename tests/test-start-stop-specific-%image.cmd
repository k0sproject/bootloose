# Test start/stop specific node in cluster

bootloose config create --override --config %testName.bootloose --name %testName --key %testName-key --image %image --replicas 3
%defer rm -f %testName.bootloose %testName-key %testName-key.pub
%defer bootloose delete --config %testName.bootloose

bootloose create --config %testName.bootloose

bootloose stop %testName-node1 --config %testName.bootloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"
%out docker inspect %testName-node1 -f "{{.State.Running}}"

bootloose start %testName-node1 --config %testName.bootloose
%out docker inspect %testName-node1 -f "{{.State.Running}}"

bootloose stop %testName-node0 --config %testName.bootloose
bootloose stop --config %testName.bootloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"
%out docker inspect %testName-node1 -f "{{.State.Running}}"
%out docker inspect %testName-node2 -f "{{.State.Running}}"
