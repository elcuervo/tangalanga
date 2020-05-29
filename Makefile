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
	@chmod +x build/*
	@echo version: $(VERSION)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -o $(BUILD_PATH)/$(WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)"  .

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -o $(BUILD_PATH)/$(LINUX) -ldflags="-s -w -X main.version=$(VERSION)"  ./

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -o $(BUILD_PATH)/$(DARWIN) -ldflags="-s -w -X main.version=$(VERSION)"  ./

clean:
	rm -f $(WINDOWS) $(LINUX) $(DARWIN)
