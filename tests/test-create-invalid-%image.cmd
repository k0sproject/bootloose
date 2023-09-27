# Test that bootloose create fails when the config is invalid (missing node number %d)

%defer bootloose delete --config test-create-invalid.static
%error bootloose create --config test-create-invalid.static
