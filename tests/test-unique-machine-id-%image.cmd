# Test that machines get unique machine-ids

footloose config create --config %testName.footloose --override --name %testName --key %testName-key --image %image
%defer rm -f %testName.footloose %testName-key %testName-key.pub
%defer footloose delete --config %testName.footloose

footloose create --config %testName.footloose
%1 footloose --config %testName.footloose ssh root@node0 -- cat /etc/machine-id 2>/dev/null || cat /var/lib/dbus/machine-id 2>/dev/null || (echo "Neither file exists" >&2; exit 1)

#%assert notempty %1
footloose delete --config %testName.footloose

footloose create --config %testName.footloose
%2 footloose --config %testName.footloose ssh root@node0 -- cat /etc/machine-id 2>/dev/null || cat /var/lib/dbus/machine-id 2>/dev/null || (echo "Neither file exists" >&2; exit 1)

%assert notequal %1 %2
