bootloose config create --override --config %testName.bootloose --name %testName --key %testName-key --image ubuntu18.04
%defer rm -f %testName.bootloose %testName-key %testName-key.pub
%defer bootloose delete --config %testName.bootloose
bootloose create --config %testName.bootloose
bootloose delete --config %testName.bootloose
%out bootloose show --config %testName.bootloose
%out bootloose show -o json --config %testName.bootloose
