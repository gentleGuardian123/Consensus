package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	PoW "github.com/gentleGuardian123/consensus/CMACProofofWork"
	PoS "github.com/gentleGuardian123/consensus/ProofofStake"
)

func CMACPoWCLI() {
	newblock := flag.NewFlagSet("newblock", flag.ExitOnError)
	listblocks := flag.NewFlagSet("listblocks", flag.ExitOnError)
	newblockData := newblock.String("data", "default data", "data for adding new block.")
	if ( len(os.Args) < 2 || !( strings.EqualFold(os.Args[1], "newblock") || strings.EqualFold(os.Args[1], "listblocks") )) {
        fmt.Println("expected 'newblock' or 'listblocks' subcommands")
        os.Exit(1)
    }
	
	fmt.Print("Enter the .db file you store blocks: ")
	reader := bufio.NewReader(os.Stdin)
	dbfile, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)	
	}
	dbfile = strings.Replace(dbfile, "\n", "", 1)
	bc := PoW.NewBlockchain(dbfile)
	fmt.Println()

	switch os.Args[1] {
	case "newblock":
        newblock.Parse(os.Args[2:])
        bc.AddBlock(*newblockData)
    case "listblocks":
        listblocks.Parse(os.Args[2:])
        bc.BlockChainIteration()
    default:
        fmt.Println("expected 'newblock' or 'listblocks' subcommands")
        os.Exit(1)
    }
}

func PoSCLI() {
	newblock := flag.NewFlagSet("newblock", flag.ExitOnError)
	newminer := flag.NewFlagSet("newminer", flag.ExitOnError)
	listblocks := flag.NewFlagSet("listblocks", flag.ExitOnError)
	listminers := flag.NewFlagSet("listminers", flag.ExitOnError)
	newblockData := newblock.String("data", "default data", "data for adding new block.")
	newminerTokens := newminer.Int64("tokens", int64(1), "tokens the miner holds with.")
	newminerDays := newminer.Int64("days", int64(1), "days the miner holds tokens for.")
	listblocksAddress := listblocks.String("address", "", "address of the miner to list blocks.")
	if ( len(os.Args) < 2 || 
	!( strings.EqualFold(os.Args[1], "newblock") || strings.EqualFold(os.Args[1], "newminer") ||
	 strings.EqualFold(os.Args[1], "listblocks") || strings.EqualFold(os.Args[1], "listminers"))) {
        fmt.Println("expected 'newblock', 'newminer', 'listblocks' or 'listminers' subcommands")
		fmt.Println("newblock:")
		newblock.PrintDefaults()
		fmt.Println("newminer:")
		newminer.PrintDefaults()
		fmt.Println("listblocks (optional):")
		listblocks.PrintDefaults()
        os.Exit(1)
    }

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the .db file you store miners and blocks: ")
	dbfile, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("Enter the .json file you store miners' addresses: ")
	addressesJson, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	dbfile = strings.Replace(dbfile, "\n", "", 1)
	addressesJson = strings.Replace(addressesJson, "\n", "", 1)

	mp := PoS.NewMinerPool(addressesJson, dbfile)
	fmt.Println()

	switch os.Args[1] {
	case "newblock":
		bc := PoS.NewBlockchain(mp)
        newblock.Parse(os.Args[2:])
        bc.AddBlock(*newblockData)
	case "newminer":
		newminer.Parse(os.Args[2:])
		mp.AddMiner(*newminerTokens, *newminerDays)
		fmt.Printf("The new added miner's address is %s, with %d tokens for %d days.", mp.Addresses[len(mp.Addresses)-1], *newminerTokens, *newminerDays)
		fmt.Println()
    case "listblocks":
		bc := PoS.NewBlockchain(mp)
        listblocks.Parse(os.Args[2:])
		if len(os.Args) >2 && strings.HasPrefix(os.Args[2], "-address=") {
			bc.ShowBlocksOfMiner(*listblocksAddress)
		} else {
			bc.BlockChainIteration()
		}
	case "listminers":
		listminers.Parse(os.Args[2:])
		mp.ShowAllMiners()
    default:
        fmt.Println("expected 'newblock', 'newminer', 'listblocks' or 'listminers' subcommands")
		newblock.PrintDefaults()
		newminer.PrintDefaults()
		listblocks.PrintDefaults()
		listminers.PrintDefaults()
        os.Exit(1)
    }
}

func main() {
	
	// CMACPoWCLI()

	PoSCLI()

}
