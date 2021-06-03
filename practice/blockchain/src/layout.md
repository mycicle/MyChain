Need to make a simple local blockchain for keeping track of transactions

Block
    Header
        Parent Hash
        Time 
        Number

    Transactions
    Nonce

Transaction
    To 
    From
    Value
    Data

Account string

State
    The state will keep track of the current state of the blockchain and will load all of the previous transactions and blocks

    latestBlock
    latestBlockHash
    balances
        Account
        Value
    
    txMempool (pending transactions)
    dbFile
    cacheFile


    NewStateFromDisk
    AddBlock
    applyTx
    verify


Node
