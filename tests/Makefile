IMAGES := $(shell $(MAKE) -s -C ../images list)

TEST_FILES := $(wildcard test-*.cmd)
TEST_NAMES := $(patsubst test-%.cmd,%,$(TEST_FILES))

# Tests that are image-agnostic (having %image in their names)
TEST_NAMES_AG := $(filter %-%image,$(TEST_NAMES))
TEST_NAMES_AG := $(patsubst %-%image,%,$(TEST_NAMES_AG))

# Tests that are image-specific (not having %image in their names)
TEST_NAMES_SP := $(filter-out %-%image,$(TEST_NAMES))

# Now we generate all possible test-image combinations
ALL_TESTS_AG := $(foreach test,$(TEST_NAMES_AG),$(foreach img,$(IMAGES),$(test)-$(img)))
ALL_TESTS_SP := $(TEST_NAMES_SP)

# Combine both for a full list
ALL_TESTS := $(ALL_TESTS_AG) $(ALL_TESTS_SP)
ALL_TESTS_GENERAL := $(TEST_NAMES_AG) $(IMAGES)
ALL_TESTS_FULL := $(addprefix test-,$(ALL_TESTS) $(ALL_TESTS_GENERAL))

comma:= ,
empty:=
space:= $(empty) $(empty)

# Test all images
.PHONY: test
test:
	go test -v -args -image="$(subst $(space),$(comma),$(IMAGES))"

# Run all tests for an individual image
# Usage: make test-ubuntu18.04
.PHONY: test-%
test-%: image-%
	go test -v -args -image=$* ./...

# Run a specific test for all images
# Usage: make test-basic-commands
define test_specific
.PHONY: test-$(1)
test-$(1): images
	go test -v -run=TestEndToEnd/test-$(1) ./ -args -image="$(subst $$(space),$$(comma),$$(IMAGES))"
endef
$(foreach test,$(TEST_NAMES_AG),$(eval $(call test_specific,$(test))))
$(foreach test,$(TEST_NAMES_SP),$(eval $(call test_specific,$(test))))

define test_specific_image_ag
.PHONY: test-$(1)-%
test-$(1)-%: image-%
	go test -v -run=TestEndToEnd/test-$(1) ./ -args -image=$$(lastword $$(subst -, ,$$(subst test-$(1)-,,$$@)))
endef
$(foreach test,$(TEST_NAMES_AG),$(eval $(call test_specific_image_ag,$(test))))

# List all tests
# Usage: make list-tests
list:
	@echo $(ALL_TESTS_FULL)

list-images:
	@echo $(IMAGES)

images:
	$(MAKE) -C ../images

define build_image
.PHONY: image-$(1)
image-$(1):
	$(MAKE) -C ../images $(1)
endef
$(foreach img,$(IMAGES),$(eval $(call build_image,$(img))))

# Phony targets
.PHONY: list list-images images
