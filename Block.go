package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64
	PrevBlockHash []byte
	Data          []byte
	Hash          []byte
	Nonce         int
}

// setHash is to hash the headers of the block which are the all the information in a block
// 把整个区块都加密了: 所以第一步先把时间变成byte格式, 然后连接所有信息组成header
func (b *Block) setHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10)) //convert int to byte
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:] // [:] bytes to slice: [32]byte -> []byte
}

func NewBlock(prevBlockHash []byte, data string) *Block {
	block := &Block{time.Now().Unix(), prevBlockHash, []byte(data), []byte{}, 0}
	// block.setHash()
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}
