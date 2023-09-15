all: footloose

footloose: bin/footloose

bin/footloose:
	go build -v -o bin/footloose .

install:
	go install -v .

# Build all images
images:
	@$(MAKE) -C images all

# Build a specific image
image-%:
	@$(MAKE) -C images $*

test-unit:
	go test -v . ./pkg/...

# Run tests against all images
test-e2e:
	@$(MAKE) -C tests test-all

# Run tests against a specific image
test-e2e-%:
	@$(MAKE) -C tests test-$*

test: test-unit test-e2e

# List available images
list-images:
	@$(MAKE) -C images list

# Clean up all stamps and other generated files
clean:
	@$(MAKE) -C images clean
	rm -f bin/footloose

# Phony targets
.PHONY: images image-% test-unit test-e2e test-e2e-% list-images clean
