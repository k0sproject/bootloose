# Test that cluster start / stop works

bootloose config create --override --config %testName.bootloose --name %testName --key %testName-key --image %image

%defer rm -f %testName.bootloose %testName-key %testName-key.pub
%defer bootloose delete --config %testName.bootloose
bootloose create --config %testName.bootloose

%out docker ps --format {{.Names}} -f label=io.k0sproject.bootloose.cluster=%testName

bootloose stop --config %testName.bootloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"

bootloose start --config %testName.bootloose
%out docker inspect %testName-node0 -f "{{.State.Running}}"
