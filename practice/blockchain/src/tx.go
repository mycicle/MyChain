package database

type Account struct {
	Owner string
}

type Tx struct {
	To    Account
	From  Account
	Value uint64
	Data  string
}
