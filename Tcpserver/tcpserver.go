package Tcpserver

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/JustSomeCarbon/bc-research/Blockchain"
	"github.com/davecgh/go-spew/spew"
)

// receives incoming blocks
var BCServer chan []Blockchain.Block

/*
 * Creates a TCP connection and waits for clients to connect to server
 */
func Connection() {
	BCServer = make(chan []Blockchain.Block)

	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on port :9000")
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}

/*
 * Handle the TCP connection for the application, requesting Orr value
 * data for new blocks
 */
func handleConn(conn net.Conn) {
	defer conn.Close()

	io.WriteString(conn, "enter Orr value: ")
	scanner := bufio.NewScanner(conn)

	// take Orr value and generate a new block for the chain
	go func() {
		for scanner.Scan() {
			orrVal, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("%v not a number: %v", scanner.Text(), err)
				continue
			}
			newBlock, err := Blockchain.GenerateBlock(Blockchain.Chain[len(Blockchain.Chain)-1], orrVal)
			if err != nil {
				log.Println(err)
				continue
			}
			if Blockchain.IsValidBlock(newBlock, Blockchain.Chain[len(Blockchain.Chain)-1]) {
				newChain := append(Blockchain.Chain, newBlock)
				Blockchain.ReplaceChain(newChain)
			}
			BCServer <- Blockchain.Chain
			io.WriteString(conn, "\nEnter Orr value: ")
		}
	}()

	// marshal and broadcast blockchain to the network
	go func() {
		for {
			time.Sleep(10 * time.Second)
			output, err := json.Marshal(Blockchain.Chain)
			if err != nil {
				log.Fatal(err)
			}
			io.WriteString(conn, string(output))
		}
	}()

	for range BCServer {
		spew.Dump(Blockchain.Chain)
	}
}
