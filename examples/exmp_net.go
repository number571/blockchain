package main

import (
	"fmt"
	"net"
	"strings"
	"./network" 
)

const (
	TITLE_OPTION = "[TITLE-MESSAGE]"
	SERVER_ADDR  = ":8080"
)

func main() {
	network.Listen(SERVER_ADDR, handleFunc)
	for i := 0; i < 3; i++ {
		pack := network.Send(SERVER_ADDR, &network.Package{
			Option: TITLE_OPTION,
			Data: "hello, world!",
		})
		fmt.Println(network.Serialize(pack))
	}
}

func handleFunc(conn net.Conn, pack *network.Package) {
	network.HandleAction(TITLE_OPTION, conn, pack, handleGet)
}

func handleGet(pack *network.Package) (set string) {
	return strings.ToUpper(pack.Data)
}
