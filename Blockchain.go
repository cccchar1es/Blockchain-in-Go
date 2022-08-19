package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

var dbFile = "blockchain.db"
var blocksBucket = "blocks"

const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
	tip []byte   // the previous hash, so we can add next blcok
	db  *bolt.DB // the db file, will be created when creating a new blockchain
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (bc *Blockchain) MineBlock(tx []*Transaction) {
	var lastHash []byte // previous hash

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	logErr(err)

	newBlock := NewBlock(lastHash, tx)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash
		logErr(err)

		return nil
	})
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]byte{}, []*Transaction{coinbase})
}

func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	logErr(err)

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	logErr(err)

	bc := Blockchain{tip, db}

	return &bc
}

func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	logErr(err)

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTx(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		logErr(err)

		err = b.Put(genesis.Hash, genesis.serialize())
		logErr(err)

		err = b.Put([]byte("l"), genesis.Hash)
		logErr(err)
		tip = genesis.Hash

		return nil
	})

	logErr(err)

	bc := Blockchain{tip, db}

	return &bc
}
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

func (i *BlockchainIterator) Next() *Block {
	var blcok *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		blcok = DeserializeBlock(encodedBlock)

		return nil
	})
	logErr(err)
	i.currentHash = blcok.PrevBlockHash
	return blcok
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		// for each transaction
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs: // for each output
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentout := range spentTXOs[txID] {
						if outIdx == spentout {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						intxId := hex.EncodeToString(in.Txid)
						spentTXOs[intxId] = append(spentTXOs[intxId], in.Vout)
					}
				}
			}

		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTX := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTX {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				accumulated += out.Value

				if accumulated > amount {
					break Work
				}
			}
		}

	}

	return accumulated, unspentOutputs
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
