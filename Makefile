.DEFAULT_GOAL := up

PROTO_DIR := ./ecommerce/proto
PROTO_FILES := $(notdir $(wildcard $(PROTO_DIR)/*.proto))

YML_DIR := ./deploy
YML_PROTO_COMPILER := proto_compiler.yml
YML_FILES := $(filter-out $(YML_PROTO_COMPILER),$(notdir $(wildcard $(YML_DIR)/*.yml)))

TARGET_SERVICES ?=

ifneq ($(strip $(TARGET_SERVICES)),)
  target_services := $(TARGET_SERVICES)
endif

# RUN
build: compile_proto
	docker compose --profile run $(foreach file,$(YML_FILES),-f $(YML_DIR)/$(file)) build $(target_services)

up:
	docker compose --profile run $(foreach file,$(YML_FILES),-f $(YML_DIR)/$(file)) up $(target_services)

up_bg:
	docker compose --profile run $(foreach file,$(YML_FILES),-f $(YML_DIR)/$(file)) up -d $(target_services)

stop:
	docker compose --profile run $(foreach file,$(YML_FILES),-f $(YML_DIR)/$(file)) stop $(target_services)

down:
	docker compose --profile run $(foreach file,$(YML_FILES),-f $(YML_DIR)/$(file)) down -v $(target_services)

# TEST
build_test: down_test compile_proto
	docker compose --profile test $(foreach file,$(YML_FILES),-f $(YML_DIR)/$(file)) build $(target_services)

test: build_test
	docker compose --profile test $(foreach file,$(YML_FILES),-f $(YML_DIR)/$(file)) up $(target_services)

down_test:
	docker compose --profile test $(foreach file,$(YML_FILES),-f $(YML_DIR)/$(file)) down -v $(target_services)

# PROTO
build_proto_image:
	docker compose -f $(YML_DIR)/$(YML_PROTO_COMPILER) build

compile_proto: build_proto_image
	docker compose -f $(YML_DIR)/$(YML_PROTO_COMPILER) run --rm protoc-builder $(PROTO_FILES)

# REMOVES DOCKER CONTAINERS, VOLUMES AND NETWORKS
purge: down down_test