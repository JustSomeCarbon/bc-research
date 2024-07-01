package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/JustSomeCarbon/bc-research/Blockchain"
	"github.com/JustSomeCarbon/bc-research/Tcpserver"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

/* -- MAIN FUNCTION -- */

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	Tcpserver.BCServer = make(chan Blockchain.Block)

	go func() {
		t := time.Now()
		genesisBlock := Blockchain.Block{Index: 0, Timestamp: t.String(), Orr: 0, Hash: "000", PrevHash: ""}
		spew.Dump(genesisBlock)
		Blockchain.Chain = append(Blockchain.Chain, genesisBlock)
	}()

	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	// run the server application
	//log.Fatal(Webserver.Run())
}
