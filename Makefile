
GO := go

SRC_DIR := src
SRC_MAIN_PATH := $(SRC_DIR)/main.go

go/fmt-golines:
	cd $(SRC_DIR) && golines --max-len=70 --tab-len=4 -w .

go/run:
	cd $(SRC_DIR) && $(GO) run .

