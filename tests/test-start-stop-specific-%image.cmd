# Test start/stop specific node in cluster

footloose config create --override --config %testName.footloose --name %testName --key %testName-key --privileged --image %image --replicas 3
%defer rm -f %testName.footloose %testName-key %testName-key.pub
%defer footloose delete --config %testName.footloose

footloose create --config %testName.footloose

footloose stop %testName-node1 --config %testName.footloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"
%out docker inspect %testName-node1 -f "{{.State.Running}}"

footloose start %testName-node1 --config %testName.footloose
%out docker inspect %testName-node1 -f "{{.State.Running}}"

footloose stop %testName-node0 --config %testName.footloose
footloose stop --config %testName.footloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"
%out docker inspect %testName-node1 -f "{{.State.Running}}"
%out docker inspect %testName-node2 -f "{{.State.Running}}"
