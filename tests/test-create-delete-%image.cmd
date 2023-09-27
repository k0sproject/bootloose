# This test creates a cluster and then deletes it

bootloose config create --override --config %testName.bootloose --name %testName --key %testName-key --image %image
%defer rm -f %testName.bootloose %testName-key %testName-key.pub
%defer bootloose delete --config %testName.bootloose
bootloose create --config %testName.bootloose
%out docker ps --format {{.Names}} -f label=io.k0sproject.bootloose.cluster=%testName
bootloose delete --config %testName.bootloose
%out docker ps --format {{.Names}} -f label=io.k0sproject.bootloose.cluster=%testName
