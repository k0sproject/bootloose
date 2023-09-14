footloose config create --override --config %testName.footloose --name %testName --key %testName-key --image %image
%defer footloose delete --config %testName.footloose
footloose create --config %testName.footloose
footloose create --config %testName.footloose
footloose delete --config %testName.footloose
footloose delete --config %testName.footloose
