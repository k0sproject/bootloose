# Test that cluster start / stop works

footloose config create --override --config %testName.footloose --name %testName --key %testName-key --privileged --image %image

%defer rm -f %testName.footloose %testName-key %testName-key.pub
%defer footloose delete --config %testName.footloose
footloose create --config %testName.footloose

%out docker ps --format {{.Names}} -f label=io.k0sproject.footloose.cluster=%testName

footloose stop --config %testName.footloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"

footloose start --config %testName.footloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"
