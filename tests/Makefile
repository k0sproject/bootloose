# Include the images Makefile to gain access to its targets and variables
include ../images/Makefile

# Test all images
test-all: $(STAMPS) install
	@go test ./ -args -image=$($(MAKE) -C ../images list IFS=,)

# Test individual images
test-%: ../images/.stamp-% install
	@go test -v -args -image=$* ./...

install:
	$(MAKE) -C .. install

# Phony targets
.PHONY: test-all list test-% install
