%defer footloose delete --config %testName.yaml
footloose create --config %testName.yaml
footloose --config %testName.yaml ssh root@node0 -- apt update && apt install -y docker.io
footloose --config %testName.yaml ssh root@node0 systemctl start docker
footloose --config %testName.yaml ssh root@node0 docker pull busybox
%out footloose --config %testName.yaml ssh root@node0 docker run busybox echo success
