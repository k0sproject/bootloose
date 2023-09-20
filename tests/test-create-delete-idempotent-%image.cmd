# This test makes sure that cluster create and delete can be run multiple times without error 
# when no changes to cluster state are necessary

footloose config create --override --config %testName.footloose --name %testName --key %testName-key --image %image
%defer footloose delete --config %testName.footloose

%defer rm -f %testName.footloose %testName-key %testName-key.pub
footloose create --config %testName.footloose
footloose create --config %testName.footloose

footloose delete --config %testName.footloose
footloose delete --config %testName.footloose
