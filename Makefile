PROJECT:=pushMessage

.PHONY: build

build:
	CGO_ENABLED=0 go build -o pushMessage main.go
test:
	go test -v ./... -cover
deps:
	go mod tidy