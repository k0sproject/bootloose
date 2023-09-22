# Tests that footloose config create fails if the config already exists

rm -f %testName.footloose
footloose config create --config %testName.footloose --name %testName --key %testName-key --privileged --image %image
%defer rm -f %testName.footloose %testName-key %testName-key.pub

%error footloose config create --config %testName.footloose --name %testName --key %testName-key --image %image
