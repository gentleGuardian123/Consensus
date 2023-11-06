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
	mp			*MinerPool
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func NewBlock(data string, prevHash []byte, height int64, mp *MinerPool) *Block {
	block := &Block{
		Timestamp: time.Now().Unix(),
		Data: []byte(data),
		PrevHash: prevHash,
		Hash: []byte{},
		Height: height,
		DiffNum: mp.DiffNum,
		Nonce: 0,
		Address: "",
	}
	SetHashAndAddress(block, mp)

	return block
}

func NewGenesisBlock(mp *MinerPool) *Block {
	fmt.Println("No existing blockchain found. Creating a new blockchain...")

	return NewBlock("Genesis Block", []byte(""), 1, mp)
}

func NewBlockchain(mp *MinerPool) *Blockchain {

	var tip []byte
	
	if err := mp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := NewGenesisBlock(mp)
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
	}); err != nil {
		log.Panic(err)
	}

	return &Blockchain{tip, mp}
}

func (bc *Blockchain) NewIterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.mp.db}
	return bci
}

func SetHashAndAddress(b *Block, mp *MinerPool) {
	if nonce, hash, address := mp.Mine(b); hash == nil {
		log.Fatal("Mine Failed!")
	} else {
		b.Nonce = nonce
		b.Hash = hash
		b.Address = address
	}
}

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte
	var lastBlock *Block

	err := bc.mp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		lastBlock = DeserializeBlock(b.Get(lastHash))
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	newBlock := NewBlock(data, lastHash, lastBlock.Height+1, bc.mp)
	err = bc.mp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, Serialize(newBlock))
		if err != nil {
			log.Fatal(err)
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Fatal(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (bci *BlockchainIterator) Next() *Block {
	var block *Block

	err := bci.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(bci.currentHash)
		block = DeserializeBlock(encodedBlock)
		bci.currentHash = block.PrevHash

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return block
}

func (bc *Blockchain) BlockChainIteration() {
	bci := bc.NewIterator()

	for {
		b := bci.Next()
		fmt.Print("Timestamp:")
		fmt.Println(time.Unix(b.Timestamp, 0).Format(time.RFC3339))
		fmt.Printf("PrevHash: %x\n", b.PrevHash)
		fmt.Printf("Data: %s\n", b.Data)
		fmt.Printf("Hash: %x\n", b.Hash)
		fmt.Printf("Height: %d\n", b.Height)
		fmt.Printf("Miner address: %s\n", b.Address)
		fmt.Printf("PoS: %t\n", bc.mp.Validate(b))
		fmt.Println()

		if b.Height == 1 {
			return
		}
	}
}

func (bc *Blockchain) ShowBlocksOfMiner(address string) {
	bci := bc.NewIterator()
	fmt.Printf("All the blocks of miner at %s is:\n", address)
	fmt.Println()

	for {
		b := bci.Next()

		if b.Address == address {
			fmt.Print("Timestamp:")
			fmt.Println(time.Unix(b.Timestamp, 0).Format(time.RFC3339))
			fmt.Printf("PrevHash: %x\n", b.PrevHash)
			fmt.Printf("Data: %s\n", b.Data)
			fmt.Printf("Hash: %x\n", b.Hash)
			fmt.Printf("Height: %d\n", b.Height)
			fmt.Printf("PoS: %t\n", bc.mp.Validate(b))
			fmt.Println()
		}
		
		if b.Height == 1 {
			return
		}
	}
}
