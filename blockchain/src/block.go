package database

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Hash [32]byte

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)

	return err
}

type Block struct {
	Header BlockHeader `json:"header"`  // metadata (parent block hash + time)
	TXs    []Tx        `json:"payload"` // new transactions only (payload)
}

type BlockHeader struct {
	Parent Hash   `json:"parent"` // parent block reference
	Number uint64 `json:"number"`
	Time   uint64 `json:"time"`
}

type BlockFS struct {
	Key   Hash  `json:"hash"`
	Value Block `json:"block"`
}

func NewBlock(parent Hash, number uint64, time uint64, txs []Tx) Block {
	return Block{
		Header: BlockHeader{
			Parent: parent,
			Number: number,
			Time:   time,
		},
		TXs: txs,
	}
}
func (b Block) Hash() (Hash, error) {
	blockJson, err := json.Marshal(b)
	if err != nil {
		return Hash{}, err
	}

	return sha256.Sum256(blockJson), nil
}

func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) IsEmpty() bool {
	emptyHash := Hash{}

	return bytes.Equal(emptyHash[:], h[:])
}
