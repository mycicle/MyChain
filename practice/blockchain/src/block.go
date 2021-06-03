package database

type Hash [32]byte

type BlockHeader struct {
	Parent Hash
	Time   uint64
	Number uint64
}

type Block struct {
	Header       BlockHeader
	Transactions Tx
	Nonce        [32]byte
}
