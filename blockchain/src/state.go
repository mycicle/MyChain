package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"
)

// state is the database component responsible for encapsulating all business logic
// it will know all user balances and who transfered TBB tokens to whom, and how many were transferred
type State struct {
	Balances        map[Account]uint
	txMempool       []Tx
	latestBlockHash Hash
	latestBlock     Block
	hasGenesisBlock bool

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
		latestBlock:     Block{},
		hasGenesisBlock: false,
		dbFile:          dbf,
		cacheFile:       cf,
	}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		blockFsJson := scanner.Bytes()

		if len(blockFsJson) == 0 {
			break
		}

		var blockFs BlockFS
		err = json.Unmarshal(blockFsJson, &blockFs)
		if err != nil {
			return nil, err
		}

		if err := applyTXs(blockFs.Value.TXs, state); err != nil {
			return nil, err
		}

		state.latestBlockHash = blockFs.Key
		state.latestBlock = blockFs.Value
		state.hasGenesisBlock = true
	}

	return state, nil
}

// Get the next block number
func (state *State) NextBlockNumber() uint64 {
	if !state.hasGenesisBlock {
		return uint64(0)
	}

	return state.latestBlock.Header.Number + 1
}

// Adding new blocks to the mempool
func (s *State) AddBlock(b Block) (Hash, error) {
	pendingState := s.copy()

	err := applyBlock(b, pendingState)
	if err != nil {
		return Hash{}, err
	}

	blockHash, err := b.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFS{
		Key:   blockHash,
		Value: b,
	}

	blockFsJson, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Printf("Persisting new Block to disk:\n")
	fmt.Printf("\t%s\n", blockFsJson)

	_, err = s.dbFile.Write(append(blockFsJson, '\n'))
	if err != nil {
		return Hash{}, err
	}

	s.Balances = pendingState.Balances
	s.latestBlockHash = blockHash
	s.latestBlock = b
	s.hasGenesisBlock = true

	return blockHash, nil
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

func (s *State) LatestBlock() Block {
	return s.latestBlock
}

// Persisting transactions to disk
func (state *State) Persist() (Hash, error) {
	block := NewBlock(
		state.latestBlockHash,
		state.latestBlock.Header.Number+1,
		uint64(time.Now().Unix()),
		state.txMempool,
	)

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

func (state *State) copy() State {
	c := State{}
	c.hasGenesisBlock = state.hasGenesisBlock
	c.latestBlock = state.latestBlock
	c.latestBlockHash = state.latestBlockHash
	c.txMempool = make([]Tx, len(state.txMempool))
	c.Balances = make(map[Account]uint)

	for acc, balance := range state.Balances {
		c.Balances[acc] = balance
	}

	for _, tx := range state.txMempool {
		c.txMempool = append(c.txMempool, tx)
	}

	return c
}

// applyBlock verifies if a block can be added to the blockchain
// Block meteada are verified as well as transactions within (sufficient balances, etc...)
func applyBlock(b Block, s State) error {
	nextExpectedBlockNumber := s.latestBlock.Header.Number + 1

	if s.hasGenesisBlock && b.Header.Number != nextExpectedBlockNumber {
		return fmt.Errorf("next expected block must be '%d' not '%d'", nextExpectedBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.latestBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'", s.latestBlockHash, b.Header.Parent)
	}

	return applyTXs(b.TXs, &s)
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

func applyTXs(txs []Tx, s *State) error {
	for _, tx := range txs {
		err := applyTx(tx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyTx(tx Tx, s *State) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf("Invalid TX. Sender '%s' balance is %d TBB. TX cost is %d TBB", tx.From, s.Balances[tx.From], tx.Value)
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}
