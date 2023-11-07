package DPoS

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
	"sort"

	"github.com/boltdb/bolt"
)


type Miner struct {
	Tokens 		int64
	Days 		int64
	Address 	string
	AddrVoted	string
	TokensPool	int64
}

type MinerPool struct {
	DiffNum		uint
	Target 		big.Int
	Addresses 	[]string
	Witnesses	[]string	
	NextWitn	int
	Probability []string
	jn 			string
	db 			*bolt.DB
}

func NewMiner(tokens, days int64) *Miner {
	return &Miner{
		Tokens: tokens,
		Days: days,
		Address: GetAddress(GenPublicKey()),
		AddrVoted: "",
		TokensPool: int64(0),
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
	mp.Witnesses = []string{}
	mp.NextWitn = 0
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

func (mp *MinerPool) Mine(b *Block) error {
	if len(mp.Witnesses) == 0 {
		log.Fatal(errors.New("there is no witnesses settled, please vote for witnesses first"))
	}

	if mp.NextWitn == witnessNum {
		log.Fatal(errors.New("the current witnesses has expired, please re-vote for witnesses."))
	}

	fmt.Printf("Mining the block containing \"%s\"\n", string(b.Data))
	fmt.Println()
	
	var hashInt big.Int
	var hash [32]byte

	for nonce := int64(0); true; nonce ++ {
		data := mp.PrepareData(b, nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(&mp.Target) < 1 {
			b.Nonce = nonce
			b.Hash = hash[:]
			b.Address = mp.Witnesses[mp.NextWitn]
			mp.NextWitn ++
			b.Witnesses = make([]string, witnessNum)
			copy(b.Witnesses, mp.Witnesses)
			return nil
		}
	}

	return errors.New("Mine Falied!")
}

func (mp *MinerPool) Validate(b *Block) bool {
	var hashInt big.Int
	hashInt.SetBytes(b.Hash)
	hashComputed := sha256.Sum256(mp.PrepareData(b, b.Nonce))
	ValidWitn := false
	for _, address := range b.Witnesses {
		if address == b.Address {
			ValidWitn = true
		}
	}

	return (hashInt.Cmp(&mp.Target) < 1) && (reflect.DeepEqual(b.Hash, hashComputed[:])) && (ValidWitn)
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

func (mp *MinerPool) Vote() {
	// Vote for one round; all miners vote (randomly) for some miner first, and can change their choice after; 
	// pick out the top ${witnessNum} ranked miners.

	if len(mp.Addresses) < witnessNum * 10 {
		log.Fatalf("The quantity of current miners is not enough; pleasure asure there are more than %d miners.", witnessNum * 10)
	}

	if err := mp.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(minerBucket))

		// Initialize the tokens pool for all miners;
		fmt.Printf("Step 1: Initializing miners' tokens pool...")
		fmt.Println()
		for _, address := range mp.Addresses {
			miner := DeserializeMiner(b.Get([]byte(address)))
			miner.AddrVoted = ""
			miner.TokensPool = int64(0)
			if err := b.Put([]byte(miner.Address), Serialize(miner)); err != nil {
				log.Fatal(err)
			}
		}

		// Vote first;
		fmt.Printf("Step 2: Voting for %d witnesses...", witnessNum)
		fmt.Println()
		for _, address := range mp.Addresses {
			miner := DeserializeMiner(b.Get([]byte(address)))
			for {
				miner.AddrVoted = mp.Addresses[rand.Intn(len(mp.Addresses))]
				if miner.Address != miner.AddrVoted {
					break
				}
			}
			minerVoted := DeserializeMiner(b.Get([]byte(miner.AddrVoted)))
			minerVoted.TokensPool += miner.Tokens * miner.Days
			if err := b.Put([]byte(miner.Address), Serialize(miner)); err != nil {
				log.Fatal(err)
			}
			if err := b.Put([]byte(minerVoted.Address), Serialize(minerVoted)); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Miner at %s voted for miner at %s, whose tokens pool is now %d.", miner.Address, minerVoted.Address, minerVoted.TokensPool)
			fmt.Println()
		}
		fmt.Println()

		// Change or hold choice after;
		fmt.Printf("Step 3: Confirming the choice...")
		fmt.Println()
		for _, address := range mp.Addresses {
			miner := DeserializeMiner(b.Get([]byte(address)))
			// HoldOrNot is ture, then continue;
			HoldOrNot := rand.Intn(2) == 1
			if HoldOrNot {
				fmt.Printf("Miner at %s holded its choice.", miner.Address)
				fmt.Println()
			} else {
				minerVotedBefore := DeserializeMiner(b.Get([]byte(miner.AddrVoted)))
				minerVotedBefore.TokensPool -= miner.Tokens * miner.Days
				for {
					miner.AddrVoted = mp.Addresses[rand.Intn(len(mp.Addresses))]
					if miner.Address != miner.AddrVoted && minerVotedBefore.Address != miner.AddrVoted {
						break
					}
				}
				minerVotedAfter := DeserializeMiner(b.Get([]byte(miner.AddrVoted)))
				minerVotedAfter.TokensPool += miner.Tokens * miner.Days
				if err := b.Put([]byte(miner.Address), Serialize(miner)); err != nil {
					log.Fatal(err)
				}
				if err := b.Put([]byte(minerVotedBefore.Address), Serialize(minerVotedBefore)); err != nil {
					log.Fatal(err)
				}
				if err := b.Put([]byte(minerVotedAfter.Address), Serialize(minerVotedAfter)); err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Miner at %s changed its choice and voted for miner at %s, whose tokens pool is now %d.", miner.Address, minerVotedAfter.Address, minerVotedAfter.TokensPool)
				fmt.Println()
			}
		}
		fmt.Println()

		// Select the top ${witnessNum} miners.
		fmt.Printf("Step 4: Deciding the final %d witnesses...", witnessNum)
		fmt.Println()
		AddressAndTokensPool := []struct {
			Address		string
			TokensPool 	int64
		}{}
		for _, address := range mp.Addresses {
			AddressAndTokensPool = append(AddressAndTokensPool, struct{Address string; TokensPool int64}{
				address,
				DeserializeMiner(b.Get([]byte(address))).TokensPool,
			})
		}
		sort.Slice(AddressAndTokensPool, func(i, j int) bool {
			return AddressAndTokensPool[i].TokensPool > AddressAndTokensPool[j].TokensPool
		})
		for i := 0; i < witnessNum; i ++ {
			mp.Witnesses = append(mp.Witnesses, AddressAndTokensPool[i].Address)
			fmt.Printf("The top %d miner is at %s, whose tokens pool is: %d.", i+1, AddressAndTokensPool[i].Address, AddressAndTokensPool[i].TokensPool)
			fmt.Println()
		}
		fmt.Println()

		// Shuffle the witnesses array.
		for i := range mp.Witnesses {
			j := rand.Intn(i + 1)
			mp.Witnesses[i], mp.Witnesses[j] = mp.Witnesses[j], mp.Witnesses[i]
		}
		
		return nil

	}); err != nil {
		log.Fatal(err)
	}

}

