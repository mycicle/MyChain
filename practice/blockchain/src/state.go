package database

import (
	"os"
)

type State struct {
	latestBlock     Block
	latestBlockHash Hash
	balances        map[Account]uint64
	txMempool       []Tx

	dbFile *os.File
}

func NewStateFromDisk(dataDir string) (State, error) {
	dir, err := initDataDirIfNotExists(dataDir)

}
