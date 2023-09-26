# List of all images
IMAGES := $(shell find . -mindepth 2 -maxdepth 2 -name Dockerfile -exec dirname {} \; | xargs -I {} basename {})

# .stamp file for each image
STAMPS := $(addsuffix /.stamp, $(IMAGES))

all: $(STAMPS)

$(foreach img, $(IMAGES), $(eval $(img)/.stamp: $(img)/Dockerfile; docker build -t $(img) $(img); touch $(img)/.stamp))

all: $(STAMPS)

# List available images
list:
	@echo $(IMAGES)

clean:
	@rm -f $(STAMPS)

%-platforms:
	@./list-image-platforms.sh $(@:-platforms=)

.PHONY: all list clean %-platforms
