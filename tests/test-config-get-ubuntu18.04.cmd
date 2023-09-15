footloose config create --override --config %testName.footloose --name %testName --key %testName-key --networks=net1,net2 --image %image
%defer rm -f %testName.footloose %testName-key %testName-key.pub
%out footloose config get --config %testName.footloose machines[0].spec
