footloose config create --override --config %testName.footloose --name %testName --key %testName-key --image %image
%defer rm -f %testName.footloose %testName-key %testName-key.pub
%defer footloose delete --config %testName.footloose
footloose create --config %testName.footloose
footloose create --config %testName.footloose
footloose delete --config %testName.footloose
footloose delete --config %testName.footloose
