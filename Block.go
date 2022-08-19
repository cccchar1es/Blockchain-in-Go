package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	PrevBlockHash []byte
	// Data          []byte
	Transactions []*Transaction
	Hash         []byte
	Nonce        int
}

// setHash is to hash the headers of the block which are the all the information in a block
// 把整个区块都加密了: 所以第一步先把时间变成byte格式, 然后连接所有信息组成header
// func (b *Block) setHash() {
// 	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10)) //convert int to byte
// 	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
// 	hash := sha256.Sum256(headers)

// 	b.Hash = hash[:] // [:] bytes to slice: [32]byte -> []byte
// }

func NewBlock(prevBlockHash []byte, Txs []*Transaction) *Block {
	block := &Block{time.Now().Unix(), prevBlockHash, Txs, []byte{}, 0}
	// block.setHash()
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (b *Block) serialize() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
