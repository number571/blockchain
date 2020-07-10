.PHONY: default build node client
default: build
build: node.go client.go serve.go
	go build -o node node.go serve.go values.go
	go build -o client client.go values.go
node: node
	./node
client: client
	./client
