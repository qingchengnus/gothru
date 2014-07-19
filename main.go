package main

import (
	"github.com/qingchengnus/gofw/bypasser"
	"net"
)

func main() {
	listenAddress, _ := net.ResolveTCPAddr("tcp", ":18888")
	ln, err := net.ListenTCP("tcp", listenAddress)
	if err != nil {

	} else {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				continue
			}
			go bypasser.HandleConnectionNegotiation(conn)
		}
	}
}
