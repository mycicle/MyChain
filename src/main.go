package main

import (
	"errors"
	"fmt"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	next     *Block
}

type BlockChain struct {
	// this is going to be a glorified linkedlist
	// so that we have O(1) addition times
	head   *Block
	tail   *Block
	length int64
}

func Genesis() *BlockChain {
	// create a new blockchain and return its address
	// emtpy head block and the tail points to the head
	return &BlockChain{
		head: &Block{
			Hash:     nil,
			Data:     nil,
			PrevHash: nil,
			next:     Genesis().head,
		},
		tail:   Genesis().head,
		length: 1,
	}
}

func (chain *BlockChain) AddBlock(block *Block) error {
	if chain.length <= 0 {
		return errors.New("Uninitialized chain used")
	}

	chain.tail.next = block
	chain.tail = block
	chain.length++

	return nil
}

func (chain *BlockChain) printChain() error {
	if chain.length <= 0 {
		return errors.New("Uninitialized chain used")
	}

	current := chain.head
	if current == nil {
		return errors.New("Head of chain is nil but the length != 0")
	}

	fmt.Printf("%+v\n", current)
	for current.next != nil {
		current = current.next
		fmt.Printf("%+v\n", current)
	}

	return nil
}

func main() {

}
