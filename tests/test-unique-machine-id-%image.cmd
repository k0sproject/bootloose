# Test that machines get unique machine-ids

bootloose config create --config %testName.bootloose --override --name %testName --key %testName-key --image %image
%defer rm -f %testName.bootloose %testName-key %testName-key.pub
%defer bootloose delete --config %testName.bootloose

bootloose create --config %testName.bootloose
%1 bootloose --config %testName.bootloose ssh root@node0 -- cat /etc/machine-id 2>/dev/null || cat /var/lib/dbus/machine-id 2>/dev/null || (echo "Neither file exists" >&2; exit 1)

#%assert notempty %1
bootloose delete --config %testName.bootloose

bootloose create --config %testName.bootloose
%2 bootloose --config %testName.bootloose ssh root@node0 -- cat /etc/machine-id 2>/dev/null || cat /var/lib/dbus/machine-id 2>/dev/null || (echo "Neither file exists" >&2; exit 1)

%assert notequal %1 %2
