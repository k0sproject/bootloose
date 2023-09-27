# Checks the output of `bootloose config get` 

bootloose config create --override --config %testName.bootloose --name %testName --key %testName-key --networks=net1,net2 --image %image
%defer rm -f %testName.bootloose %testName-key %testName-key.pub

%out bootloose config get --config %testName.bootloose machines[0].spec
