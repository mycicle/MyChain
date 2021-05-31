package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// state is the database component responsible for encapsulating all business logic
// it will know all user balances and who transfered TBB tokens to whom, and how many were transferred
type State struct {
	Balances        map[Account]uint
	txMempool       []Tx
	latestBlockHash Hash

	dbFile    *os.File
	cacheFile *os.File
}

// it is contstructed using the initial balances from the genesis.json file
func NewStateFromDisk(dataDir string) (*State, error) {
	err := initDataDirIfNotExists(dataDir)
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(getGenesisJsonFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	dbf, err := os.OpenFile(getBlocksDbFilePath(dataDir), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	cf, err := os.OpenFile(getStateJsonFilePath(dataDir), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(dbf)

	state := &State{
		Balances:        balances,
		txMempool:       make([]Tx, 0),
		latestBlockHash: Hash{},
		dbFile:          dbf,
		cacheFile:       cf,
	}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		blockFsJson := scanner.Bytes()
		var blockFs BlockFS
		err = json.Unmarshal(blockFsJson, &blockFs)
		if err != nil {
			return nil, err
		}

		if err := state.applyBlock(blockFs.Value); err != nil {
			return nil, err
		}

		state.latestBlockHash = blockFs.Key
	}

	return state, nil
}

// Adding new blocks to the mempool
func (state *State) AddBlock(block Block) error {
	for _, tx := range block.TXs {
		if err := state.AddTx(tx); err != nil {
			return err
		}
	}

	return nil
}

// Adding new transactions to the mempool
func (state *State) AddTx(tx Tx) error {
	if err := state.apply(tx); err != nil {
		return err
	}

	state.txMempool = append(state.txMempool, tx)

	return nil
}

func (state *State) LatestBlockHash() Hash {
	return state.latestBlockHash
}

// Persisting transactions to disk
func (state *State) Persist() (Hash, error) {
	block := NewBlock(state.latestBlockHash, uint64(time.Now().Unix()), state.txMempool)
	blockHash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFS{
		Key:   blockHash,
		Value: block,
	}

	blockFsJson, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Printf("Persisting new block to disk:\n")
	fmt.Printf("\t%s\n", blockFsJson)

	if _, err = state.dbFile.Write(append(blockFsJson, '\n')); err != nil {
		return Hash{}, err
	}

	state.latestBlockHash = blockHash
	state.txMempool = []Tx{}

	return state.latestBlockHash, nil
}

func (state *State) Close() {
	state.cacheFile.Close()
	state.dbFile.Close()
}

func (state *State) applyBlock(b Block) error {
	for _, tx := range b.TXs {
		if err := state.apply(tx); err != nil {
			return err
		}
	}

	return nil
}

// Validating transactions against the current state (valid sender balance)
// Changing the state
func (state *State) apply(tx Tx) error {
	if tx.IsReward() {
		state.Balances[tx.To] += tx.Value
		return nil
	}

	if state.Balances[tx.From] < tx.Value {
		return fmt.Errorf("insufficient balance")
	}

	state.Balances[tx.From] -= tx.Value
	state.Balances[tx.To] += tx.Value

	return nil
}
