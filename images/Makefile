# List of all images
IMAGES := $(shell find . -mindepth 2 -maxdepth 2 -name Dockerfile -exec dirname {} \; | xargs -I {} basename {})

# .stamp file for each image
STAMPS := $(addsuffix /.stamp, $(IMAGES))

all: $(STAMPS)

# generate rules
$(foreach img, $(IMAGES), $(eval $(img)/.stamp: $(img)/Dockerfile; docker build -t $(img) $(img); touch $(img)/.stamp))

# Default target
all: $(STAMPS)

# List available images
list:
	@echo $(IMAGES)

# Clean up
clean:
	@rm -f $(STAMPS)

# Phony targets
.PHONY: all list clean
