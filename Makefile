EXECUTABLE=tangalanga
WINDOWS=build/$(EXECUTABLE)_windows_amd64.exe
LINUX=build/$(EXECUTABLE)_linux_amd64
DARWIN=build/$(EXECUTABLE)_darwin_amd64
VERSION=$(shell git describe --tags --always --long --dirty)

.PHONY: all clean

all: build

windows: $(WINDOWS)

linux: $(LINUX)

darwin: $(DARWIN)

build: windows linux darwin
	@echo version: $(VERSION)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -o $(WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)"  .

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -o $(LINUX) -ldflags="-s -w -X main.version=$(VERSION)"  ./

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -o $(DARWIN) -ldflags="-s -w -X main.version=$(VERSION)"  ./

clean:
	rm -f $(WINDOWS) $(LINUX) $(DARWIN)
