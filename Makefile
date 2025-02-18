all: gen-proto build
start: debug
	cd bin && ./backend 2>&1 | tee /dev/tty | multilog s10485760 n100 ./log

OS := $(shell uname -s)

gen-proto:
ifeq ($(OS), Darwin)
	@echo "Generating proto files for macOS"
	make gen-proto-unix
else ifeq ($(OS), Linux)
	@echo "Generating proto files for Linux"
	make gen-proto-unix
else
	@echo "Generating proto files for Windows"
	make gen-proto-win
endif

gen-proto-unix:
	rm -rf generated
	mkdir generated
	protoc -Iproto/api -Iproto/third_party --go_out=paths=source_relative:generated --go-grpc_out=paths=source_relative:generated --go-db-enum_out=paths=source_relative:generated proto/api/**/*.proto

gen-proto-win:
	rd grpc/generated
	mkdir grpc/generated
	cmd /c gen-proto.bat
	protoc -Iproto/api -Iproto/third_party --go_out=paths=source_relative:generated --go-grpc_out=paths=source_relative:generated --go-db-enum_out=paths=source_relative:generated proto/api/**/*.proto


debug:
ifeq ($(OS),Darwin)
	@echo "Building in debug mode for MacOS"
	make gen-proto build-mac-debug
else ifeq ($(OS),Linux)
	@echo "Building in debug mode for Linux"
	make gen-proto build-linux-debug
else
	@echo "Building in debug mode for Windows"
	make gen-proto build-win-debug
endif

build:
ifeq ($(OS),Darwin)
	@echo "Building for MacOS"
	make gen-proto build-mac
else ifeq ($(OS),Linux)
	@echo "Building for Linux"
	make gen-proto build-linux
else
	@echo "Building for Windows"
	make gen-proto build-win
endif

build-win:
	GOOS=windows GOARCH=amd64 go build -o bin/backend.exe cmd/main.go

build-win-debug:
	GOOS=windows GOARCH=amd64 go build -tags debug -gcflags='all=-N -l' -o bin/backend.exe cmd/main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/backend cmd/main.go

build-linux-debug:
	GOOS=linux GOARCH=amd64 go build -tags debug -gcflags='all=-N -l' -o bin/backend cmd/main.go

build-mac:
	GOOS=darwin GOARCH=arm64 go build -o bin/backend cmd/main.go

build-mac-debug:
	GOOS=darwin GOARCH=arm64 go build -tags debug -gcflags='all=-N -l' -o bin/backend cmd/main.go