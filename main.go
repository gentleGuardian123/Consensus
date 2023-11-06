package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	PoW "github.com/gentleGuardian123/consensus/CMACProofofWork"
)


func main() {

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
		log.Panic(err)	
		os.Exit(1)
	}
	dbfile = strings.Replace(dbfile, "\n", "", -1)
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
