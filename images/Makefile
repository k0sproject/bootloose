IFS ?= "\n"

# List of all images
IMAGES := $(shell find . -mindepth 1 -maxdepth 1 -type d -exec basename {} \;)

# .stamp file for each image
STAMPS = $(addprefix .stamp-, $(IMAGES))

# Default target
all: $(STAMPS)

# Rule to build an individual image
.stamp-%: %/Dockerfile
	docker build -t $(notdir $*) ./$*
	@touch $*/.stamp

# List available images
list:
	@echo $(subst $(eval) ,$(IFS),$(IMAGES))

# Clean up
clean:
	@rm -f .stamp-*

# Phony targets
.PHONY: all list clean

# Make each stamp file dependent on its Dockerfile
$(foreach img, $(IMAGES), $(eval .stamp-$(img): $(img)/Dockerfile))
