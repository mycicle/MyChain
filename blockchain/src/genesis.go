package database

import (
	"encoding/json"
	"io/ioutil"
)

const genesisJson = `
{
    "genesis_time": "2021-05-26T00:00:00.000000000Z",
    "chain_id": "the-blockchain-bar-ledger",
    "balances": {
        "andrej": 1000000
    }
}
`

type genesis struct {
	Balances map[Account]uint `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return genesis{}, err
	}

	var loadedGenesis genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return genesis{}, err
	}

	return loadedGenesis, nil
}

func writeGenesisToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(genesisJson), 0644)
}
