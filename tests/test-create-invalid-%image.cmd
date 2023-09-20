# Test that footloose create fails when the config is invalid (missing node number %d)

%defer footloose delete --config test-create-invalid.static
%error footloose create --config test-create-invalid.static
