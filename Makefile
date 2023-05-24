LDFLAGS := -s -w

all: fmt clean build

build: svr cli

svr:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/linux-server ./server
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/osx-server ./server
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/windows-server ./server

cli:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/linux-client ./client
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/osx-client ./client
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/windows-client ./client

clean:
	rm -f ./bin/*-client
	rm -f ./bin/*-server

fmt:
	go fmt ./...
