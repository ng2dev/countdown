package client

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"bitbucket.org/swoldt/pkg/xerrors/iferr"
	countdown "github.com/ng2dev/countdown/cmd/countdown/app"
	"github.com/iov-one/weave"
	weaveClient "github.com/iov-one/weave/client"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/commands/server"
	"github.com/iov-one/weave/crypto"
	"github.com/iov-one/weave/migration"
	"github.com/iov-one/weave/x/cash"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	rpctest "github.com/tendermint/tendermint/rpc/test"
	tm "github.com/tendermint/tendermint/types"
)

// configuration for genesis
var initBalance = coin.Coin{
	Whole:  100200300,
	Ticker: "COUNTDOWN",
}

// adjust this to get debug output
var logger = log.NewNopLogger()

// useful values for test cases
var node *nm.Node
var faucet *crypto.PrivateKey

func getChainID() string {
	return rpctest.GetConfig().ChainID()
}

func TestMain(m *testing.M) {
	faucet = GenPrivateKey()

	config := rpctest.GetConfig()
	config.Moniker = "SetInTestMain"

	// set up our application
	admin := faucet.PublicKey().Address()
	app, err := initApp(config, admin)
	iferr.Exit(err)

	code := weaveClient.TestWithTendermint(app, func(n *nm.Node) {
		node = n
	}, m)

	os.Exit(code)
}

func initApp(config *cfg.Config, addr weave.Address) (abci.Application, error) {
	opts := &server.Options{
		MinFee: coin.Coin{},
		Home:   config.RootDir,
		Logger: logger,
		Debug:  false,
	}

	app, err := countdown.GenerateApp(opts)
	if err != nil {
		return nil, err
	}

	// generate genesis file...
	err = initGenesis(config.GenesisFile(), addr)

	return app, err
}

func initGenesis(filename string, addr weave.Address) error {
	doc, err := tm.GenesisDocFromFile(filename)
	if err != nil {
		return err
	}
	appState, err := json.Marshal(map[string]interface{}{
		"cash": []interface{}{
			dict{
				"address": addr,
				"coins":   coin.Coins{&initBalance},
			},
		},
		"conf": dict{
			"cash": cash.Configuration{
				CollectorAddress: weave.NewAddress([]byte("fake-collector-address")),
				MinimalFee:       coin.Coin{}, // no fee
			},
			"migration": migration.Configuration{
				Admin: weave.Condition("multisig/usage/0000000000000001").Address(),
			},
		},
		"initialize_schema": []dict{
			{"pkg": "migration", "ver": 1},
			{"pkg": "cash", "ver": 1},
			{"pkg": "sigs", "ver": 1},
			{"pkg": "multisig", "ver": 1},
			{"pkg": "utils", "ver": 1},
			{"pkg": "validators", "ver": 1},
		},
	})
	if err != nil {
		return fmt.Errorf("serialize state: %s", err)
	}
	doc.AppState = appState
	return doc.SaveAs(filename)
}

type dict map[string]interface{}
