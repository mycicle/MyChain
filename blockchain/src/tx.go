package database

// each customer in the database is represented by an account struct
type Account string

func NewAccount(value string) Account {
	return Account(value)
}

// each transaction has a from, to, value, and data
type Tx struct {
	From  Account `json:"from"`
	To    Account `json:"to"`
	Value uint    `json:"value"`
	Data  string  `json:"data"`
}

func NewTx(from Account, to Account, value uint, data string) Tx {
	return Tx{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	}
}

// if we are spawning new tokens to reward someone then the data field is set to reward
func (t Tx) IsReward() bool {
	return t.Data == "reward"
}
