footloose config create --config %testName.footloose --name %testName --key %testName-key --image %image
%defer rm -f %testName.footloose %testName-key %testName-key.pub
footloose config create --config %testName.footloose --name %testName --key %testName-key --image %image
