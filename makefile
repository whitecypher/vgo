XFLAGS = -X main.version=$(shell git describe --tags)

install:
    @go install -ldflags "-w ${XFLAGS}" ./main.go
    // @ go build -o ./bin/wecure-server -ldflags "-w ${XFLAGS}" app/server/main.go

