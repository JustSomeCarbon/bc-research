package Webserver

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JustSomeCarbon/bc-research/Blockchain"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
)

/*
 * A struct to hold message information
 */
type Message struct {
	Orr int
}

/*
 * runs the server at the specified port, returns an error
 * if something goes wrong, nil else
 */
func Run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("PORT")
	log.Println("Listening on: ", httpAddr)
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
	bytes, err := json.MarshalIndent(Blockchain.Chain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	io.WriteString(w, string(bytes))
}

/*
 * parses the request body for message information to build a new block for
 * the blockchain. if an error occurs, a response is sent to the client
 * and the function returns
 */
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJson(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock, err := Blockchain.GenerateBlock(Blockchain.Chain[len(Blockchain.Chain)-1], m.Orr)
	if err != nil {
		respondWithJson(w, r, http.StatusInternalServerError, m)
		return
	}

	if Blockchain.IsValidBlock(newBlock, Blockchain.Chain[len(Blockchain.Chain)-1]) {
		newBlockChain := append(Blockchain.Chain, newBlock)
		Blockchain.ReplaceChain(newBlockChain)
		spew.Dump(Blockchain.Chain)
	}

	respondWithJson(w, r, http.StatusCreated, newBlock)
}

func respondWithJson(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}
