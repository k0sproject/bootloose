%defer bootloose delete --config %testName.yaml
bootloose create --config %testName.yaml
bootloose --config %testName.yaml ssh root@node0 -- amazon-linux-extras install -y docker
bootloose --config %testName.yaml ssh root@node0 systemctl start docker
bootloose --config %testName.yaml ssh root@node0 docker pull busybox
%out bootloose --config %testName.yaml ssh root@node0 docker run busybox echo success
