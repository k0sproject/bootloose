# Tests that bootloose config create fails if the config already exists

rm -f %testName.bootloose
bootloose config create --config %testName.bootloose --name %testName --key %testName-key --image %image
%defer rm -f %testName.bootloose %testName-key %testName-key.pub

%error bootloose config create --config %testName.bootloose --name %testName --key %testName-key --image %image
