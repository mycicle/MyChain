package main

import (
	"fmt"
	"os"
	"time"

	database "github.com/mycicle/MyChain/blockchain/src"
)

func main() {
	cwd, _ := os.Getwd()

	state, err := database.NewStateFromDisk(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer state.Close()

	// manually transfer the first set of transactions to blocks
	block0 := database.NewBlock(
		database.Hash{},
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("andrej", "andrej", 3, ""),
			database.NewTx("andrej", "andrej", 700, "reward"),
		},
	)

	state.AddBlock(block0)
	block0hash, err := state.Persist()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Manually transfer the next set of transactions to blocks
	block1 := database.NewBlock(
		block0hash,
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("andrej", "babayaga", 2000, ""),
			database.NewTx("andrej", "andrej", 100, "reward"),
			database.NewTx("babayaga", "andrej", 1, ""),
			database.NewTx("babayaga", "caesar", 1000, ""),
			database.NewTx("babayaga", "andrej", 50, ""),
			database.NewTx("andrej", "andrej", 600, "reward"),
		},
	)

	state.AddBlock(block1)
	_, err = state.Persist()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
