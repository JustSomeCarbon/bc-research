package Blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	Orr       int
	Hash      string
	PrevHash  string
}

var Chain []Block

/*
 * calculate the hash of the given block and return the hash as a string Orrue
 */
func CalculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.Orr) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

/*
 * generate a new block for the blockchain, return the new block
 */
func GenerateBlock(oldBlock Block, Orr int) (Block, error) {
	var newBlock Block
	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Orr = Orr
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = CalculateHash(newBlock)

	return newBlock, nil
}

/*
 * determines if a block is valid in the blockchain, returns true if it
 * is, false otherwise.
 */
func IsValidBlock(block Block, oldBlock Block) bool {
	if oldBlock.Index+1 != block.Index {
		return false
	}
	if oldBlock.Hash != block.PrevHash {
		return false
	}
	if CalculateHash(block) != block.Hash {
		return false
	}
	return true
}

/*
 * determines which chain is the right one, sets the local
 * blockchain to the longest given chain
 */
func ReplaceChain(newChain []Block) {
	if len(newChain) > len(Chain) {
		Chain = newChain
	}
}
