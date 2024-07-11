package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JustSomeCarbon/bc-research/Blockchain"
	_ "github.com/JustSomeCarbon/bc-research/Tcpserver"

	"github.com/davecgh/go-spew/spew"
	golog "github.com/ipfs/go-log"
	"github.com/joho/godotenv"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p/core/crypto" // ../core/[package]
	host "github.com/libp2p/go-libp2p/core/host"
	net "github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	gologging "github.com/whyrusleeping/go-logging"
)

/* -- MAIN FUNCTION -- */

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Tcpserver.BCServer = make(chan Blockchain.Block)

	go func() {
		t := time.Now()
		genesisBlock := Blockchain.Block{Index: 0, Timestamp: t.String(), Orr: 0, Hash: "000", PrevHash: ""}
		spew.Dump(genesisBlock)
		Blockchain.Chain = append(Blockchain.Chain, genesisBlock)
	}()

	/* p2p application implementation */

	/* use golog to log messages. They log with different string ID's (ex: "swarm")
	 * We control the verbosity level for the loggers with:
	 */
	golog.SetAllLoggers(golog.LogLevel(gologging.INFO))

	// parse the options from the command line
	listenF := flag.Int("l", 0, "wait for incomming conenctions")
	target := flag.String("d", "", "target peer dial")
	secio := flag.Bool("secio", false, "enable secio (deprecated)")
	seed := flag.Int64("seed", 0, "set random number seed for id gen")
	// parse the given arguments and fill the defined flags
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// make a host that listens on the given multiaddress
	ha, err := MakeBasicHost(*listenF, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}

	// look for target connections
	if *target == "" {
		log.Println("listening for connections...")
		// set a stream handler on the host. "/p2p/1.0.0" is
		// a user-defined protocol name
		ha.SetStreamHandler("p2p/1.0.0", HandleStream)

		select {} // hang forever
		/* end of listener code */

	} else {
		ha.SetStreamHandler("p2p/1.0.0", HandleStream)

		// the following code extracts the target's peer ID from
		// the given multiaddress
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		// extract the peer id from the multiaddress
		peerInfo, err := peer.AddrInfoFromP2pAddr(ipfsaddr)
		if err != nil {
			log.Fatal(err)
		}

		// add the peer multiaddress to the peerstore.
		// this will be used during connection and stream creation
		// by libp2p.
		ha.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)

		log.Println("starting a stream with peer...")

		// start a new stream with the destination
		// multiaddress of the destination peer is fetched from the peerstore using the 'peerId'
		s, err := ha.NewStream(context.Background(), peerInfo.ID, "/p2p/1.0.0")
		if err != nil {
			log.Fatal(err)
		}

		// create a buffered stream so that the read and writes are not blocking
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		// create threads to read and write data
		go WriteData(rw)
		go ReadData(rw)

		select {} // wait forever
	}
}

/*
 * creates a Libp2p host with a random peer ID listening on the given
 * multiaddress. It will use secio if secio is true
 */
func MakeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {
	// If the seed is zero, use real cryptographic randomness, otherwise, use
	// deterministic randomness source to make generated keys stay the same
	// accross multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. Used to obtain a valid host ID
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	// define options for the libp2p constructor
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	if !secio {
		//opts = append(opts, libp2p.NoEncryption())
		log.Printf("no encryption flagged")
	}

	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	// build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().String()))

	// now build full multiaddress to reach this host
	// by encapsulating both addresses
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr) //.String() ?

	return basicHost, nil
}

/*
 * handles the p2p connection stream. Creates two go routines to read and write
 * from a given net stream.
 */
func HandleStream(s net.Stream) {
	log.Println("Got a new stream!")

	// create a buffer stream for non-blocking read and write
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ReadData(rw)
	go WriteData(rw)
}

/*
 * takes a pointer to a bufio ReadWriter and reads the data from the stream
 */
func ReadData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}

		if str != "\n" {
			newChain := make([]Blockchain.Block, 0)
			if err := json.Unmarshal([]byte(str), &newChain); err != nil {
				log.Fatal(err)
			}

			// Obtain the lock for the blockchain
			Blockchain.Mutex.Lock()
			if len(newChain) > len(Blockchain.Chain) {
				Blockchain.Chain = newChain
				bytes, err := json.MarshalIndent(Blockchain.Chain, "", " ")
				if err != nil {
					log.Fatal(err)
				}

				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			// unlock the lock for the blockchain
			Blockchain.Mutex.Unlock()
		}
	}
}

/*
 * takes a pointer to a bufio ReadWriter and writes data to the stream
 */
func WriteData(rw *bufio.ReadWriter) {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			// lock the blockchain
			Blockchain.Mutex.Lock()
			bytes, err := json.Marshal(Blockchain.Chain)
			if err != nil {
				log.Fatal(err)
			}
			// unlock the blockchain
			Blockchain.Mutex.Unlock()

			// lock the blockchain
			Blockchain.Mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			// unlock the blockchain
			Blockchain.Mutex.Unlock()
		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("?> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		orr, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}

		newBlock, err := Blockchain.GenerateBlock(Blockchain.Chain[len(Blockchain.Chain)-1], orr)

		if Blockchain.IsValidBlock(newBlock, Blockchain.Chain[len(Blockchain.Chain)-1]) {
			Blockchain.Mutex.Lock()
			Blockchain.Chain = append(Blockchain.Chain, newBlock)
			Blockchain.Mutex.Unlock()
		}

		bytes, err := json.Marshal(Blockchain.Chain)
		if err != nil {
			log.Println(err)
		}

		spew.Dump(Blockchain.Chain)

		Blockchain.Mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		Blockchain.Mutex.Unlock()
	}
}
