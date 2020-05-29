EXECUTABLE=tangalanga
BUILD_PATH=build
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
LINUX=$(EXECUTABLE)_linux_amd64
DARWIN=$(EXECUTABLE)_darwin_amd64
VERSION=$(shell git describe --tags --always --long --dirty)

.PHONY: all clean proto

all: build

windows: $(WINDOWS)

linux: $(LINUX)

darwin: $(DARWIN)

proto:
	protoc --go_out=proto meeting.proto

build: windows linux darwin
	@echo version: $(VERSION)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -o $(BUILD_PATH)/$(WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)"  .
	@chmod +x $(BUILD_PATH)/$(WINDOWS)
	zip -r $(BUILD_PATH)/$(WINDOWS).zip $(BUILD_PATH)/$(WINDOWS)

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -o $(BUILD_PATH)/$(LINUX) -ldflags="-s -w -X main.version=$(VERSION)"  ./
	@chmod +x $(BUILD_PATH)/$(LINUX)
	tar cfz $(BUILD_PATH)/$(LINUX).tgz $(BUILD_PATH)/$(LINUX)

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -o $(BUILD_PATH)/$(DARWIN) -ldflags="-s -w -X main.version=$(VERSION)"  ./
	@chmod +x $(BUILD_PATH)/$(DARWIN)
	tar cfz $(BUILD_PATH)/$(DARWIN).tgz $(BUILD_PATH)/$(DARWIN)

clean:
	rm -f $(BUILD_PATH)/*
