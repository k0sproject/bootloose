footloose config create --override --config %testName.footloose --name %testName --key %testName-key --image %image
%defer footloose delete --config %testName.footloose
footloose create --config %testName.footloose
%out docker ps --format {{.Names}} -f label=io.k0sproject.footloose.cluster=%testName
footloose stop --config %testName.footloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"
footloose start --config %testName.footloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"
