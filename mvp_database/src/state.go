package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// state is the database component responsible for encapsulating all business logic
// it will know all user balances and who transfered TBB tokens to whom, and how many were transferred
type State struct {
	Balances  map[Account]uint
	txMempool []Tx

	dbFile    *os.File
	cacheFile *os.File
}

// it is contstructed using the initial balances from the genesis.json file
func NewStateFromDisk() (*State, error) {
	// get cwd
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	genFilePath := filepath.Join(cwd, "..", "MyChain", "mvp_database", "database", "genesis.json")
	gen, err := loadGenesis(genFilePath)
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	dbf, err := os.OpenFile(filepath.Join(cwd, "..", "MyChain", "mvp_database", "database", "tx.db"), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	cf, err := os.OpenFile(filepath.Join(cwd, "..", "MyChain", "mvp_database", "database", "state.json"), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(dbf)

	state := &State{
		Balances:  balances,
		txMempool: make([]Tx, 0),
		dbFile:    dbf,
		cacheFile: cf,
	}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var tx Tx
		err = json.Unmarshal(scanner.Bytes(), &tx)
		if err != nil {
			return nil, err
		}

		if err := state.apply(tx); err != nil {
			return nil, err
		}
	}

	return state, nil
}

// Adding new transactions to the mempool
func (state *State) Add(tx Tx) error {
	if err := state.apply(tx); err != nil {
		return err
	}

	state.txMempool = append(state.txMempool, tx)

	return nil
}

// Persisting transactions to disk
func (state *State) Persist() error {
	mempool := make([]Tx, len(state.txMempool))
	copy(mempool, state.txMempool)

	for i := 0; i < len(mempool); i++ {
		txJson, err := json.Marshal(mempool[i])
		if err != nil {
			return err
		}

		if _, err := state.dbFile.Write(append(txJson, '\n')); err != nil {
			return err
		}
		state.txMempool = state.txMempool[1:]
	}

	balancesJson, err := json.Marshal(state.Balances)
	if err != nil {
		return err
	}

	_, err = state.cacheFile.Write(balancesJson)
	if err != nil {
		return err
	}

	return nil
}

func (state *State) Close() {
	state.dbFile.Close()
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
