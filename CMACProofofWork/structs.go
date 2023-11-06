package CMACPoW

import (
	"bytes"
	"crypto/aes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"time"

	"github.com/aead/cmac"
	"github.com/boltdb/bolt"
)

const blocksBucket = "blocks"
const defaultDatabase = "blocks.db"
const difficulty = 8

type Block struct {
	Timestamp 	int64
	PrevHash 	[]byte
	Hash 		[]byte
	Data 		[]byte
	Height 		int64
	DiffNum 	uint
	Nonce 		int64
	CMAC 		[]byte
}

type Blockchain struct { tip []byte
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


func NewBlock(data string, prevHash []byte, height int64) *Block {
	block := &Block{
		Timestamp: 	time.Now().Unix(), 
		Data: 		[]byte(data), 
		PrevHash: 	prevHash,
		Hash: 		[]byte{},
		Height: 	height,	
		DiffNum: 	difficulty,
		Nonce: 		0,
		CMAC: 		[]byte{},
	}
	block.SetHash()

	return block
}

func NewGenesisBlock() *Block {
	fmt.Println("No existing blockchain found. Creating a new blockchain...")
	zeroHash := []byte("l")
	for i:= 1; i < 32; i++ {
		zeroHash = append(zeroHash, 0)
	}
	return NewBlock("Genesis Block", zeroHash, 1)
}

func NewBlockchain(dbFile string) *Blockchain {
	if len(dbFile) == 0 {
		dbFile = defaultDatabase
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Fatal(err)
			}
			err = b.Put(genesis.Hash, Serialize(genesis))
			if err != nil {
				log.Fatal(err)
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Fatal(err)
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(aes.BlockSize*8-b.DiffNum))

	return &ProofOfWork{b, target}
}

func (bc *Blockchain) NewIterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}


func (b *Block) SetHash() {
	pow := NewProofOfWork(b)
	nonce, hash, Cmac := pow.Mine()
	if hash == nil {
		log.Fatal("Mine Failed!")
	} else {
		b.Nonce = nonce
		b.Hash = hash
		b.CMAC = Cmac
	}
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
		log.Fatal(err)
	}
	newBlock := NewBlock(data, lastHash, lastBlock.Height+1)
	err = bc.db.Update(func(tx *bolt.Tx) error {
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
		log.Fatal(err)
	}

	return block
}

func (pow *ProofOfWork) Mine() (int64,[]byte, []byte) {
	fmt.Printf("Mining the block containing \"%s\"\n\n", string(pow.block.Data))

	var nonce int64
	var cmacInt big.Int
	block, err := aes.NewCipher(pow.block.PrevHash)
	if err != nil {
		log.Fatal(err)
	}
	for nonce = 1; true; nonce ++ { 
		// use cbc-mac to achieve pseudo random used for collision.
		// use nonce as plaintext and compute CBC-MAC of it.
		_nonce := make([]byte, 8)
		binary.LittleEndian.PutUint64(_nonce, uint64(nonce))
		theCmac, err := cmac.Sum(_nonce, block, block.BlockSize())
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(theCmac)
		cmacInt.SetBytes(theCmac[:])
		if cmacInt.Cmp(pow.target) < 1 {
			blockHash := sha256.Sum256(pow.PrepareData(nonce))
			return nonce, blockHash[:], theCmac[:]
		}
	}

	return -1, nil, nil
}

func (pow *ProofOfWork) Validate() bool {
	var cmacInt, theCmacInt big.Int

	block, err := aes.NewCipher(pow.block.PrevHash)
	if err != nil {
		log.Fatal(err)	
	}
	
	_nonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(_nonce, uint64(pow.block.Nonce))
	ret, err := cmac.Sum(_nonce, block, block.BlockSize())
	if err != nil {
		log.Fatal(err)
	}
	cmacInt.SetBytes(ret)
	theCmacInt.SetBytes(pow.block.CMAC)

	hashComputed := sha256.Sum256(pow.PrepareData(pow.block.Nonce))
	return (cmacInt.Cmp(&theCmacInt) == 0) && (cmacInt.Cmp(pow.target) < 1) && (reflect.DeepEqual(pow.block.Hash, hashComputed[:]))
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
		fmt.Printf("CMAC: %x\n", block.CMAC)
		fmt.Printf("PoW: %t\n", NewProofOfWork(block).Validate())
		fmt.Println()

		if block.Height == 1 {
			return
		}
	}
}