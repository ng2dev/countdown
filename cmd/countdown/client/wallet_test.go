package client

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/weavetest/assert"
	"github.com/iov-one/weave/x/cash"
)

var defaults = coin.Coin{
	Ticker:     "COUNTDOWN",
	Whole:      123456789,
	Fractional: 5555555,
}

func toWeaveAddress(t *testing.T, addr string) weave.Address {
	d, err := hex.DecodeString(addr)
	assert.Nil(t, err)
	return d
}

func wsFromFile(t *testing.T, wsFile string) (*WalletStore, error) {
	w := WalletStore{}
	err := w.LoadFromFile(wsFile, defaults)
	assert.Nil(t, err)

	wsJson, err := ToJsonString(w)
	if err != nil {
		return nil, err
	}

	t.Log(wsJson)

	return &w, nil
}

func wsFromJSON(t *testing.T, ws json.RawMessage) (*WalletStore, error) {
	w := WalletStore{}
	err := w.LoadFromJSON(ws, defaults)
	assert.Nil(t, err)

	wsJson, err := ToJsonString(w)
	if err != nil {
		return nil, err
	}
	t.Log(wsJson)

	return &w, nil
}

func wsFromGenesisFile(t *testing.T, wsFile string) (*WalletStore, error) {
	w := WalletStore{}
	err := w.LoadFromGenesisFile(wsFile, defaults)
	assert.Nil(t, err)

	wsJson, err := ToJsonString(w)
	if err != nil {
		return nil, err
	}
	t.Log(wsJson)

	return &w, nil
}

func TestMergeWalletStore(t *testing.T) {
	w1, err := wsFromGenesisFile(t, "./testdata/genesis.json")
	assert.Nil(t, err)
	w2, err := wsFromFile(t, "./testdata/wallets.json")
	assert.Nil(t, err)

	expected := WalletStore{
		Wallets: []cash.GenesisAccount{
			{
				Address: toWeaveAddress(t, "3AFCDAB4CFBF066E959D139251C8F0EE91E99D5A"),
				Set: cash.Set{
					Coins: []*coin.Coin{
						{
							Ticker:     "COUNTDOWN",
							Whole:      123456789,
							Fractional: 5555555,
						},
					},
				},
			},
			{
				Address: toWeaveAddress(t, "12AFFBF6012FD2DF21416582DC80CBF1EFDF2460"),
				Set: cash.Set{
					Coins: []*coin.Coin{
						{
							Ticker:     "COUNTDOWN",
							Whole:      987654321,
							Fractional: 5555555,
						},
					},
				},
			},
			{
				Address: toWeaveAddress(t, "E28AE9A6EB94FC88B73EB7CBD6B87BF93EB9BEF0"),
				Set: cash.Set{
					Coins: []*coin.Coin{
						{
							Ticker:     "COUNTDOWN",
							Whole:      987654321,
							Fractional: 5555555,
						},
					},
				},
			},
			{
				Address: toWeaveAddress(t, "CE5D5A5CA8C7D545D7756D3677234D81622BA297"),
				Set: cash.Set{
					Coins: []*coin.Coin{
						{
							Ticker:     "COUNTDOWN",
							Whole:      123456789,
							Fractional: 5555555,
						},
					},
				},
			},
			{
				Address: toWeaveAddress(t, "D4821FD051696273D09E1FBAD0EBE5B5060787A7"),
				Set: cash.Set{
					Coins: []*coin.Coin{
						{
							Ticker:     "COUNTDOWN",
							Whole:      123456789,
							Fractional: 5555555,
						},
					},
				},
			},
		},
	}

	actual := MergeWalletStore(*w1, *w2)
	assert.Equal(t, expected, actual)
}

func TestMergeWithEmptyWallet(t *testing.T) {
	w1, err := wsFromJSON(t, []byte(`{}`))
	assert.Nil(t, err)
	w2, err := wsFromFile(t, "./testdata/wallets.json")
	assert.Nil(t, err)

	expected := WalletStore{
		Wallets: []cash.GenesisAccount{
			{
				Address: toWeaveAddress(t, "CE5D5A5CA8C7D545D7756D3677234D81622BA297"),
				Set: cash.Set{
					Coins: []*coin.Coin{
						{
							Ticker:     "COUNTDOWN",
							Whole:      123456789,
							Fractional: 5555555,
						},
					},
				},
			},
			{
				Address: toWeaveAddress(t, "D4821FD051696273D09E1FBAD0EBE5B5060787A7"),
				Set: cash.Set{
					Coins: []*coin.Coin{
						{
							Ticker:     "COUNTDOWN",
							Whole:      123456789,
							Fractional: 5555555,
						},
					},
				},
			},
		},
	}

	actual := MergeWalletStore(*w1, *w2)
	assert.Equal(t, expected, actual)
}

func TestKeyGen(t *testing.T) {
	useCases := map[string]struct {
		W string
		N int
	}{
		"empty":  {`{}`, 0},
		"single": {`{"cash":[{}]}`, 1},
	}

	for testName, useCase := range useCases {
		t.Run(testName, func(t *testing.T) {
			w, err := wsFromJSON(t, []byte(useCase.W))
			assert.Nil(t, err)
			assert.Equal(t, useCase.N, len(w.Keys))
		})
	}
}
