package Tcpserver

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"

	"github.com/JustSomeCarbon/bc-research/Blockchain"
)

// receives incoming blocks
var BCServer chan Blockchain.Block

func Connection() {
	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	io.WriteString(conn, "enter Orr value: ")
	scanner := bufio.NewScanner(conn)

	// take Orr value and generate a new block for the chain
	go func() {
		for scanner.Scan() {
			//
		}
	}()
}
