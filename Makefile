.PHONY: default build run
default: build run
build: main.go
	go build main.go
run: main
	./main
