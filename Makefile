.PHONY: default build
default: build
build: node.go client.go serve.go values.go
	go build -o node node.go serve.go values.go
	go build -o client client.go values.go
