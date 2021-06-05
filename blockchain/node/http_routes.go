package node

import (
	"net/http"

	database "github.com/mycicle/MyChain/blockchain/src"
)

type ErrRes struct {
	Error string `json:"error"`
}

type BalancesRes struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type TxAddReq struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type TxAddRes struct {
	Hash database.Hash `json:"block_hash"`
}

type StatusRes struct {
	Hash       database.Hash       `json:"block_hash"`
	Number     uint64              `json:"block_number"`
	KnownPeers map[string]PeerNode `json:"peers_known"`
}

func listBalancesHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	writeRes(w, BalancesRes{
		Hash:     state.LatestBlockHash(),
		Balances: state.Balances,
	})
}

func txAddHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	req := TxAddReq{}
	err := readReq(r, &req)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	tx := database.NewTx(
		database.NewAccount(req.From),
		database.NewAccount(req.To),
		req.Value,
		req.Data,
	)

	// Add a new TX to the mempool
	err = state.AddTx(tx)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	// Flush the mempool to the disk
	hash, err := state.Persist()
	if err != nil {
		writeErrRes(w, err)
		return
	}

	writeRes(w, TxAddRes{
		Hash: hash,
	})

}

func statusHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	res := StatusRes{
		Hash:       node.state.LatestBlockHash(),
		Number:     node.state.LatestBlock().Header.Number,
		KnownPeers: node.knownPeers,
	}

	writeRes(w, res)
}
