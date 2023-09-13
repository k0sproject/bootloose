footloose config create --override --config %testName.footloose --name %testName --key %testName-key --image %image
footloose create --config %testName.footloose
%out footloose --config %testName.footloose ssh root@node0 whoami
footloose delete --config %testName.footloose
