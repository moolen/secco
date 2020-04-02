
binary: bin
	GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o bin/secco main.go

bin:
	mkdir bin
