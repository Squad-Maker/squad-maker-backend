all: gen-proto build-linux
debug: gen-proto build-linux-debug

start: debug
	cd bin && ./backend 2>&1 | tee /dev/tty | multilog s10485760 n100 ./log

gen-proto:
	rm -rf generated
	mkdir generated
	protoc -Iproto/api -Iproto/third_party --go_out=paths=source_relative:generated --go-grpc_out=paths=source_relative:generated --go-db-enum_out=paths=source_relative:generated proto/api/**/*.proto

build-win:
	GOOS=windows GOARCH=amd64 go build -o bin/backend.exe cmd/main.go

build-win-debug:
	GOOS=windows GOARCH=amd64 go build -tags debug -gcflags='all=-N -l' -o bin/backend.exe cmd/main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/backend cmd/main.go

build-linux-debug:
	GOOS=linux GOARCH=amd64 go build -tags debug -gcflags='all=-N -l' -o bin/backend cmd/main.go
