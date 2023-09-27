# This test makes sure that cluster create and delete can be run multiple times without error 
# when no changes to cluster state are necessary

bootloose config create --override --config %testName.bootloose --name %testName --key %testName-key --image %image
%defer bootloose delete --config %testName.bootloose

%defer rm -f %testName.bootloose %testName-key %testName-key.pub
bootloose create --config %testName.bootloose
bootloose create --config %testName.bootloose

bootloose delete --config %testName.bootloose
bootloose delete --config %testName.bootloose
