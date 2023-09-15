footloose config create --override --config %testName.footloose --name %testName --key %testName-key --image %image
%defer rm -f %testName.footloose %testName-key %testName-key.pub
%defer footloose delete --config %testName.footloose
footloose create --config %testName.footloose
footloose show --config %testName.footloose
footloose show --config %testName.footloose -o json
%out footloose --config %testName.footloose ssh root@node0 whoami
