package PoS

import (
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

type Block struct {
	Timestamp 	int64
	PrevHash 	[]byte
	Hash 		[]byte
	Data 		[]byte
	Height 		int64
	DiffNum 	uint
	Nonce 		int64
	Address 	string	
}

type Blockchain struct { 
	tip 		[]byte
	db  		*bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func NewBlock(data, address string, prevHash []byte, height int64) *Block {
	block := &Block{
		Timestamp: time.Now().Unix(),
		Data: []byte(data),
		PrevHash: prevHash,
		Hash: []byte{},
		Height: height,
		DiffNum: difficulty,
		Nonce: 0,
		Address: address,
	}
	// block.SetHash()

	return block
}

func NewGenesisBlock() *Block {
	fmt.Println("No existing blockchain found. Creating a new blockchain...")

	noStakeMiner := NewMiner(0, 0)

	return NewBlock("Genesis Block", noStakeMiner.Address, []byte(""), 1)
}

func NewBlockchain(dbFile string) *Blockchain {
	if len(dbFile) == 0 {
		dbFile = defaultDatabase
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(genesis.Hash, Serialize(genesis))
			if err != nil {
				log.Panic(err)
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}
