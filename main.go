package main

import (
	"log"
	"time"

	"github.com/JustSomeCarbon/bc-research/Blockchain"
	"github.com/JustSomeCarbon/bc-research/Webserver"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

/* -- MAIN FUNCTION -- */

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Blockchain.Block{Index: 0, Timestamp: t.String(), Orr: 0, Hash: "000", PrevHash: ""}
		spew.Dump(genesisBlock)
		Blockchain.Chain = append(Blockchain.Chain, genesisBlock)
	}()

	// run the server application
	log.Fatal(Webserver.Run())
}
