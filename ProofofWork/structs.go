package PoW

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"time"

	"github.com/boltdb/bolt"
)

const blocksBucket = "blocks"
const difficulty = 24

type Block struct {
	Timestamp int64
	PrevHash []byte
	Hash []byte
	Data []byte
	Height int
	DiffNum uint
	Nonce int64
}

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func NewBlock(data string, prevHash []byte, height int) *Block {
	block := &Block{
		Timestamp: time.Now().Unix(), 
		// Timestamp: 1567578876,
		Data: []byte(data), 
		PrevHash: prevHash,
		Hash: []byte{},
		Height: height,	
		DiffNum: difficulty,
		Nonce: 0,
	}
	block.SetHash()

	return block
}

func NewGenesisBlock() *Block {
	fmt.Println("No existing blockchain found. Creating a new blockchain...")
	return NewBlock("Genesis Block", []byte{}, 1)
}

func NewBlockchain(dbFile string) *Blockchain {
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

			err = b.Put(genesis.Hash, genesis.Serialize())
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

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-b.DiffNum))

	return &ProofOfWork{b, target}
}

func (bc *Blockchain) NewIterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}


func (b *Block) SetHash() {
	pow := NewProofOfWork(b)
	nonce, hash := pow.Mine()
	if hash == nil {
		log.Panic("Mine Failed!")
	} else {
		b.Nonce = nonce
		b.Hash = hash
	}
}

func (b *Block) Serialize() []byte {
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

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte
	var lastBlock *Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		lastBlock = DeserializeBlock(b.Get(lastHash))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(data, lastHash, lastBlock.Height+1)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		i.currentHash = block.PrevHash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return block
}

func (pow *ProofOfWork) PrepareData(nonce int64) []byte {
	return bytes.Join(
		[][]byte{
			pow.block.PrevHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(pow.block.DiffNum)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
}

func (pow *ProofOfWork) Mine() (int64, []byte) {
	fmt.Printf("Mining the block containing \"%s\"\n", string(pow.block.Data))

	var nonce int64
	var hashInt big.Int
	var hash [32]byte

	for nonce = 0; true; nonce ++ {
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) < 1 {
			return nonce, hash[:]
		}
	}

	return -1, nil
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	hashInt.SetBytes(pow.block.Hash)
	hashComputed := sha256.Sum256(pow.PrepareData(pow.block.Nonce))
	return (hashInt.Cmp(pow.target) < 1) && (reflect.DeepEqual(pow.block.Hash, hashComputed[:]))
}

func (bc *Blockchain) BlockChainIteration() {
	bci := bc.NewIterator()
	for {
		block := bci.Next()
		fmt.Print("Timestamp:")
		fmt.Println(time.Unix(block.Timestamp, 0).Format(time.RFC3339))
		fmt.Printf("PrevHash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("PoW: %t\n", NewProofOfWork(block).Validate())
		fmt.Println()

		if block.Height == 1 {
			return
		}
	}
}