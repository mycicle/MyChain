package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	database "github.com/mycicle/MyChain/blockchain/src"
)

const DefaultIP = "127.0.0.1"
const DefaultHTTPort = uint64(8080)

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`

	// Whenever my node has already established a connection, sync with this Peer
	connected bool
}

func (pn PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}

type Node struct {
	dataDir string
	ip      string
	port    uint64

	// To inject the State into HTTP handlers
	state *database.State

	knownPeers map[string]PeerNode
}

func New(dataDir string, ip string, port uint64, bootstrap PeerNode) *Node {
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TcpAddress()] = bootstrap

	return &Node{
		dataDir:    dataDir,
		ip:         ip,
		port:       port,
		knownPeers: knownPeers,
	}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, connected bool) PeerNode {
	return PeerNode{
		IP:          ip,
		Port:        port,
		IsBootstrap: isBootstrap,
		connected:   connected,
	}
}

func (n *Node) Run() error {
	// ctx := context.Background()
	fmt.Println(fmt.Sprintf("Listening on %s:%d", n.ip, n.port))

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	n.state = state

	// go n.sync(ctx)

	// GET endpoint to get the balances of everyone on the network
	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	// POST endpoint to add new transactions to the ledger
	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, state)
	})

	//GET endpoint to get the status of the node
	http.HandleFunc("/node/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}

// Run() method loads the state from disk and then starts listening on the hardcoded port 8080
// func Run(dataDir string) error {
// 	state, err := database.NewStateFromDisk(dataDir)
// 	if err != nil {
// 		return err
// 	}

// 	defer state.Close()

// 	// GET endpoint to get the balances of everyone on the network
// 	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
// 		listBalancesHandler(w, r, state)
// 	})

// 	// POST endpoint to add new transactions to the ledger
// 	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
// 		txAddHandler(w, r, state)
// 	})

// 	return http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)
// }

func writeRes(w http.ResponseWriter, content interface{}) {
	contentJson, err := json.Marshal(content)
	if err != nil {
		writeErrRes(w, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(contentJson)
}

func writeErrRes(w http.ResponseWriter, err error) {
	jsonErrRes, _ := json.Marshal(ErrRes{Error: err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsonErrRes)
}

func readReq(r *http.Request, reqBody interface{}) error {
	reqBodyJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body %s", err.Error())
	}
	defer r.Body.Close()

	err = json.Unmarshal(reqBodyJson, reqBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal request body %s", err.Error())
	}

	return nil
}
