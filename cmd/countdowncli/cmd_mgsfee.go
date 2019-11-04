package main

import (
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/x/msgfee"
)

// msgfeeConf returns message fee from blockchain
func msgfeeConf(nodeURL string, msgPath string) (*coin.Coin, error) {
	store := tendermintStore(nodeURL)
	b := msgfee.NewMsgFeeBucket()
	return b.MessageFee(store, msgPath)
}
