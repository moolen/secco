
binary: bin
	GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o bin/secco main.go

.PHONY: proto
proto:
	protoc -I proto proto/agent.proto --go_out=plugins=grpc:./proto

bin:
	mkdir bin
