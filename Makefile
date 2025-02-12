
GO := go
PYTHON := python3

SRC_DIR := src
SRC_MAIN_PATH := $(SRC_DIR)/main.go

PROXY_ARGS += -to
PROXY_ARGS += localhost:8000

SRV_DIR := srv
SRV_CMD := cd $(SRV_DIR) && $(PYTHON) -m http.server -b 0.0.0.0 8000 --directory .

go/fmt-golines:
	cd $(SRC_DIR) && golines --max-len=80 --tab-len=4 -w .

go/run:
	cd $(SRC_DIR) && $(GO) run . $(PROXY_ARGS)

net/serve:
	$(SRV_CMD)

