BINARY_NAME=backend
BUILD_DIR=./build/

run:
	go build -o ${BUILD_DIR}${BINARY_NAME} main.go
	./build/${BINARY_NAME}

build:
	go build -o ${BUILD_DIR}${BINARY_NAME} main.go

clean:
	go clean
	rm -r ./build
