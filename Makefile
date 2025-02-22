.PHONY: help go/fmt-golines go/run/server go/run/client go/build/all net/serve

GO := go
PYTHON := python3

SRC_DIR := src
SRC_MAIN_PATH := $(SRC_DIR)/main.go
SRC_BUILD_BIN_NAME := fp
SRC_RELEASE_CMD := mkdir -p release/ && cd $(SRC_DIR) && $(GO) build -o $(SRC_BUILD_BIN_NAME) && $(GO) build -ldflags "-s -w" -trimpath -o fp .
SRC_RELEASE_DIR := release
SRC_RELEASE_COPY_CMD := cp -v $(SRC_DIR)/$(SRC_BUILD_BIN_NAME) $(SRC_RELEASE_DIR)/$(SRC_BUILD_BIN_NAME)

# not used anymore
PROXY_ARGS += -to
PROXY_ARGS += localhost:8000

SRV_DIR := srv
SRV_CMD := cd $(SRV_DIR) && $(PYTHON) -m http.server -b 0.0.0.0 8000 --directory .

default: help

# shows this help list
help:
	@printf '\tMakefile targets\n\n'
	@awk ' \
		function trim(s) { \
			sub(/^[ \t]+/, "", s); \
			sub(/[ \t]+$$/, "", s); \
			return s; \
		} \
		/^$$/ { help = ""; next } \
		/^#/ { \
			help = (help ? help " " trim(substr($$0,2)) : trim(substr($$0,2))); \
			next; \
		} \
		/^[^ \t]+:/ { \
			target = $$1; sub(/:$$/, "", target); \
			if (help != "") { \
				printf "  %-15s %s\n", target, help; \
			} \
			help = ""; \
		}' $(MAKEFILE_LIST)
	@printf "\n"

# formats all go code with golines
go/fmt-golines:
	cd $(SRC_DIR) && golines --max-len=80 --tab-len=2 -w .

# runs server
go/run/server:
	cd $(SRC_DIR) && $(GO) run main.go -server

# runs client
go/run/client:
	cd $(SRC_DIR) && $(GO) run main.go -client

# builds code and outputs to executable file
go/build/all:
	cd $(SRC_DIR) && $(GO) build -o $(SRC_BUILD_BIN_NAME) main.go

# builds for release and copies executable to release directory
go/release:
	$(SRC_RELEASE_CMD)
	$(SRC_RELEASE_COPY_CMD)

# runs http server with python
net/serve:
	$(SRV_CMD)

