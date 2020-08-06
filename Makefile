.PHONY: default xbuild ybuild
default: xbuild ybuild
# Self-written part
xbuild: node.go client.go gclient.go serve.go values.go
	go build -o node node.go serve.go values.go
	go build -o client client.go values.go
	go build -o gclient gclient.go values.go
# Ethereum part
ybuild: contract.sol deploy.go client_eth.go gclient_eth.go values_eth.go
	solc --overwrite --abi --bin contract.sol -o build
	mkdir -p contracts
	./abigen --bin=./build/WorldSkills.bin --abi=./build/WorldSkills.abi --pkg=contract --out=./contracts/Contract.go
	go build -o deploy deploy.go
	go build -o client_eth client_eth.go values_eth.go
	go build -o gclient_eth gclient_eth.go values_eth.go
