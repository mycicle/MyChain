package database

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// used to store hash information after each transaction
type Snapshot [32]byte

// state is the database component responsible for encapsulating all business logic
// it will know all user balances and who transfered TBB tokens to whom, and how many were transferred
type State struct {
	Balances  map[Account]uint
	txMempool []Tx
	snapshot  Snapshot

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

	cf, err := os.OpenFile(filepath.Join(cwd, "..", "MyChain", "mvp_database", "database", "state.json"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(dbf)

	state := &State{
		Balances:  balances,
		txMempool: make([]Tx, 0),
		snapshot:  Snapshot{},
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

	err = state.doSnapshot()
	if err != nil {
		return nil, err
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

func (state *State) LatestSnapshot() Snapshot {
	return state.snapshot
}

// Persisting transactions to disk
func (state *State) Persist() (Snapshot, error) {
	mempool := make([]Tx, len(state.txMempool))
	copy(mempool, state.txMempool)

	for i := 0; i < len(mempool); i++ {
		txJson, err := json.Marshal(mempool[i])
		if err != nil {
			return Snapshot{}, err
		}

		fmt.Printf("Persisting new TX to disk:\n")
		fmt.Printf("\t%s\n", txJson)
		if _, err := state.dbFile.Write(append(txJson, '\n')); err != nil {
			return Snapshot{}, err
		}

		err = state.doSnapshot()
		if err != nil {
			return Snapshot{}, err
		}
		fmt.Printf("New DB Snapshot: %x\n", state.snapshot)

		state.txMempool = state.txMempool[1:]
	}

	balancesJson, err := json.Marshal(state.Balances)
	if err != nil {
		return Snapshot{}, err
	}

	_, err = state.cacheFile.Write(balancesJson)
	if err != nil {
		return Snapshot{}, err
	}

	return state.snapshot, nil
}

func (state *State) Close() {
	state.cacheFile.Close()
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

func (state *State) doSnapshot() error {
	// Re-read the whole file from the first byte
	_, err := state.dbFile.Seek(0, 0)
	if err != nil {
		return err
	}

	txsData, err := ioutil.ReadAll(state.dbFile)
	if err != nil {
		return err
	}

	state.snapshot = sha256.Sum256(txsData)

	return nil
}
