BINARY_NAME=backend
BUILD_DIR=./build/
SWAG_BIN=$(shell go env GOPATH)/bin/swag

build_run: build swag start

start:
	./build/${BINARY_NAME}

build: clean
	go build -o ${BUILD_DIR}${BINARY_NAME} main.go

swag: swag-install
	$(SWAG_BIN) init

swag-install:
	@if ! [ -x "$(SWAG_BIN)" ]; then \
		echo "swag not found, installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi

clean:
	go clean
	-rm -rf ./docs
	-rm -rf ./build
