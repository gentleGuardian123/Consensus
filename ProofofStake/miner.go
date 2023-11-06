package PoS

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"reflect"

	"github.com/boltdb/bolt"
)


type Miner struct {
	Tokens 		int64
	Days 		int64
	Address 	string
}

type MinerPool struct {
	DiffNum		uint
	Target 		big.Int
	Addresses 	[]string			
	Probability []string
	jn 			string
	db 			*bolt.DB
}

func NewMiner(tokens, days int64) *Miner {
	return &Miner{
		Tokens: tokens,
		Days: days,
		Address: GetAddress(GenPublicKey()),
	}
}

func NewMinerPool(addressesJson, dbFile string) *MinerPool {

	if len(dbFile) == 0 {
		dbFile = defaultDatabase
	}
	if len(addressesJson) == 0 {
		addressesJson = defaultAddressesJson
	}

	var mp MinerPool

	// if the json file exist, read from it; else create it and initialize the mp.Addresses empty.
	if _, err := os.Stat(addressesJson); err == nil {
		if _ret, err := os.ReadFile(addressesJson); err != nil {
			log.Fatal(err)
		} else {
			if err := json.Unmarshal(_ret, &mp.Addresses); len(_ret) != 0 && err != nil {
				log.Fatal(err)
			} 
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if _, err = os.Create(addressesJson); err != nil {
			log.Fatal(err)
		}
		mp.Addresses = []string{};
	} else {
		log.Fatal(err)
	}

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}	
	if err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(minerBucket))
		if b == nil {
			if b, err = tx.CreateBucket([]byte(minerBucket)); err != nil {
				log.Fatal(err)
			}
		}
		for _, address := range mp.Addresses {
			if _miner := b.Get([]byte(address)); _miner == nil {
				log.Fatalf("Unvalid address: %s", address)
			} else {
				miner := DeserializeMiner(_miner)
				for it := int64(0); it < miner.Days * miner.Tokens / stakeScare; it ++ {
					mp.Probability = append(mp.Probability, address)
				}
			}
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	mp.DiffNum = difficulty
	mp.Target = *big.NewInt(1)
	mp.Target.Lsh(&mp.Target, uint(256 - mp.DiffNum))
	mp.jn = addressesJson
	mp.db = db
	return &mp
}

func (mp *MinerPool) AddMiner(tokens, days int64) {

	miner := NewMiner(tokens, days)

	mp.Addresses = append(mp.Addresses, miner.Address)

	for it := int64(0); it < miner.Days * miner.Tokens / stakeScare; it ++ {
		mp.Probability = append(mp.Probability, miner.Address)
	}

	// write new addresses set to json file 
	if ret, err := json.MarshalIndent(mp.Addresses, "", " "); err != nil {
		log.Fatal(err)
	} else {
		if err = os.WriteFile(mp.jn, ret, 0644); err != nil {
			log.Fatal(err)
		}
	}

	// put (miner.Address, miner) into the minerbucket of dbfile
	if err := mp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(minerBucket))
		if err := b.Put([]byte(miner.Address), Serialize(miner)); err != nil {
			log.Fatal(err)
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

func (mp *MinerPool) PrepareData(b *Block, nonce int64) []byte {
	return bytes.Join(
		[][]byte{
			b.PrevHash,
			b.Data,
			IntToHex(b.Timestamp),
			IntToHex(int64(b.DiffNum)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
}

func (mp *MinerPool) Mine(b *Block) (int64, []byte, string) {
	fmt.Printf("Mining the block containing \"%s\"\n", string(b.Data))
	
	var hashInt big.Int
	var hash [32]byte

	for nonce := int64(0); true; nonce ++ {
		data := mp.PrepareData(b, nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(&mp.Target) < 1 {
			return nonce, hash[:], mp.Probability[rand.Intn(len(mp.Probability))]
		}
	}

	return -1, nil, ""
}

func (mp *MinerPool) Validate(b *Block) bool {
	var hashInt big.Int
	hashInt.SetBytes(b.Hash)
	hashComputed := sha256.Sum256(mp.PrepareData(b, b.Nonce))
	return (hashInt.Cmp(&mp.Target) < 1) && (reflect.DeepEqual(b.Hash, hashComputed[:]))
}

func (mp *MinerPool) Adjust(NewDiffNum uint) {
	mp.DiffNum = NewDiffNum
	mp.Target = *big.NewInt(1)
	mp.Target.Lsh(&mp.Target, uint(256 - mp.DiffNum))
}

func (mp *MinerPool) ShowAllMiners() {
	if err := mp.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(minerBucket))
		for _, address := range mp.Addresses {
			miner := DeserializeMiner(b.Get([]byte(address)))
			fmt.Printf("Address: %s\n", miner.Address)
			fmt.Printf("Tokens: %d\n", miner.Tokens)
			fmt.Printf("Days: %d\n", miner.Days)
			fmt.Println()
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
