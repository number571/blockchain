package network

import (
	"net"
	"time"
	"strings"
)

type Listener net.Listener
type Conn net.Conn

func Listen(address string, handle func(Conn, *Package)) Listener {
	splited := strings.Split(address, ":")
	if len(splited) != 2 {
		return nil
	}
	listener, err := net.Listen("tcp", "0.0.0.0:"+splited[1])
	if err != nil {
		return nil
	}
	go serve(listener, handle)
	return Listener(listener)
}

func serve(listener net.Listener, handle func(Conn, *Package)) {
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		go handleConn(conn, handle)
	}
}

func handleConn(conn net.Conn, handle func(Conn, *Package)) {
	defer conn.Close()
	pack := readPackage(conn)
	if pack == nil {
		return
	}
	handle(Conn(conn), pack)
}

func Send(address string, pack *Package) *Package {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil
	}
	defer conn.Close()
	conn.Write([]byte(SerializePackage(pack) + ENDBYTES))
	var res = new(Package)
 	ch := make(chan bool)
	go func() {
		res = readPackage(conn)
		ch <- true
	}()
 	select {
  		case <-ch:
  		case <-time.After(WAITTIME * time.Second):
 	}
 	return res
}

func Handle(option int, conn Conn, pack *Package, handle func(*Package) string) bool {
	if pack.Option != option {
		return false
	}
	conn.Write([]byte(SerializePackage(&Package{
		Option: option,
		Data:   handle(pack),
	}) + ENDBYTES))
	return true
}

func readPackage(conn net.Conn) *Package {
	var (
		data   string
		size   = uint64(0)
		buffer = make([]byte, BUFFSIZE)
	)
	for {
		length, err := conn.Read(buffer)
		if err != nil {
			return nil
		}
		size += uint64(length)
		if size > DMAXSIZE {
			return nil
		}
		data += string(buffer[:length])
		if strings.Contains(data, ENDBYTES) {
			data = strings.Split(data, ENDBYTES)[0]
			break
		}
	}
	return DeserializePackage(data)
}
