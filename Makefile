.PHONY: default xbuild ybuild
default: xbuild ybuild
# Self-written part
xbuild:
	go build -o node ./cmd/node
	go build -o client ./cmd/client
	go build -o gclient ./cmd/gclient
# Ethereum part
ybuild: contract.sol
	solc --overwrite --abi --bin contract.sol -o build
	mkdir -p contracts
	./abigen --bin=./build/WorldSkills.bin --abi=./build/WorldSkills.abi --pkg=contract --out=./contracts/Contract.go
	go build -o deploy ./cmd/deploy
	go build -o client_eth ./cmd/client_eth
	go build -o gclient_eth ./cmd/gclient_eth
