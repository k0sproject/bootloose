# Include the images Makefile to gain access to its targets and variables
include ../images/Makefile

# Test all images
test-all: $(STAMPS) install
	@go test ./ -args -image=$($(MAKE) -C ../images list IFS=,)

# Test individual images
test-%: ../images/.stamp-%
	@go test -v -args -image=$* ./...

# Phony targets
.PHONY: test-all list test-%
