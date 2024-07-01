package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv"
)

type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

var BlockChain []Block

/*
 * calculate the hash of the given block and return the hash as a string value
 */
func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

/*
 * generate a new block for the blockchain, return the new block
 */
func generateBlock(oldBlock Block, BPM int) (Block, error) {
	var newBlock Block
	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

/*
 * determines if a block is valid in the blockchain, returns true if it
 * is, false otherwise.
 */
func isValidBlock(block Block, oldBlock Block) bool {
	if oldBlock.Index+1 != block.Index {
		return false
	}
	if oldBlock.Hash != block.PrevHash {
		return false
	}
	if calculateHash(block) != block.Hash {
		return false
	}
	return true
}

/*
 * determines which chain is the right one, sets the local
 * blockchain to the longest given chain
 */
func replaceChain(newChain []Block) {
	if len(newChain) > len(BlockChain) {
		BlockChain = newChain
	}
}

/* -- WEB SERVER -- */

/*
 * A struct to hold message information
 */
type Message struct {
	BPM int
}

/*
 * runs the server at the specified port, returns an error
 * if something goes wrong, nil else
 */
func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on: ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

/*
 * create the gorilla/mux router used by the server.
 * returns the mux router
 */
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlock).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

/*
 * the function handler for the get request. Writes the Blockchain
 * as a marshalled value to the response writter
 */
func handleGetBlock(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(BlockChain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	io.WriteString(w, string(bytes))
}

/*
 *
 */
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		return
	}
}
