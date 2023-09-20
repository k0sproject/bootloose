# TODO: not sure what this actually proves

footloose config create --override --config %testName.footloose --name %testName --key %testName-key --image %image

%defer footloose delete --config %testName.footloose
%defer rm -f %testName.footloose %testName-key %testName-key.pub
footloose create --config %testName.footloose

%out docker ps --format {{.Names}} -f label=io.k0sproject.footloose.cluster=%testName
%out docker inspect %testName-node0 -f "{{.HostConfig.AutoRemove}}"
footloose delete --config %testName.footloose
%out docker ps --format {{.Names}} -f label=io.k0sproject.footloose.cluster=%testName
