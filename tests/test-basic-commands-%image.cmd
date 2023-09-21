# Test that common utilities are present in the base images

footloose config create --config %testName.footloose --override --name %testName --key %testName-key --image %image
%defer rm -f %testName.footloose %testName-key %testName-key.pub
%defer footloose delete --config %testName.footloose
footloose create --config %testName.footloose

%1 footloose --config %testName.footloose ssh root@node0 hostname
%assert equal %1 node0

# test uname and capture buffer interpolation
footloose --config %testName.footloose ssh root@%1 -- uname -a

footloose --config %testName.footloose ssh root@node0 ps
footloose --config %testName.footloose ssh root@node0 ifconfig
footloose --config %testName.footloose ssh root@node0 ip route
footloose --config %testName.footloose ssh root@node0 -- netstat -n -l
footloose --config %testName.footloose ssh root@node0 -- command -v ping
footloose --config %testName.footloose ssh root@node0 -- curl --version
footloose --config %testName.footloose ssh root@node0 -- command -v wget
footloose --config %testName.footloose ssh root@node0 -- vi --help
footloose --config %testName.footloose ssh root@node0 -- sudo -n true || doas -n true
