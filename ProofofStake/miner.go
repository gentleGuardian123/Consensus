package PoS

import (
	"encoding/json"
	"log"
	"os"

	"github.com/boltdb/bolt"
)


type Miner struct {
	Tokens 		int64
	Days 		int64
	Address 	string
}

type MinerPool struct {
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

	//TODO: IF addressesJosn NOT EXIST! CREATE IT FOR USER.
	if _ret, err := os.ReadFile(addressesJson); err != nil {
		log.Fatal(err)
	} else {
		if err := json.Unmarshal(_ret, &mp.Addresses); len(_ret) != 0 && err != nil {
			log.Fatal(err)
		} 
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
