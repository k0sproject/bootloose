# TODO: not sure what this actually proves

bootloose config create --override --config %testName.bootloose --name %testName --key %testName-key --image %image

%defer bootloose delete --config %testName.bootloose
%defer rm -f %testName.bootloose %testName-key %testName-key.pub
bootloose create --config %testName.bootloose

%out docker ps --format {{.Names}} -f label=io.k0sproject.bootloose.cluster=%testName
%out docker inspect %testName-node0 -f "{{.HostConfig.AutoRemove}}"
bootloose delete --config %testName.bootloose
%out docker ps --format {{.Names}} -f label=io.k0sproject.bootloose.cluster=%testName
