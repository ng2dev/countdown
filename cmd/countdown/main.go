package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/ng2dev/countdown"
	countdown "github.com/ng2dev/countdown/cmd/countdown/app"
	"github.com/iov-one/weave/commands"
	"github.com/iov-one/weave/commands/server"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	flagHome = "home"
	varHome  *string
)

func init() {
	defaultHome := filepath.Join(os.ExpandEnv("$HOME"), ".countdown")
	varHome = flag.String(flagHome, defaultHome, "directory to store files under")

	flag.CommandLine.Usage = helpMessage
}

func helpMessage() {
	fmt.Println("countdown")
	fmt.Println("          Countdown Application Blockchain Service node")
	fmt.Println("")
	fmt.Println("help      Print this message")
	fmt.Println("init      Initialize app options in genesis file")
	fmt.Println("start     Run the abci server")
	fmt.Println("getblock  Extract a block from blockchain.db")
	fmt.Println("retry     Run last block again to ensure it produces same result")
	fmt.Println("testgen   Generate various protoc and json files to test against")
	fmt.Println("version   Print the app version")
	fmt.Println(`
  -home string
        directory to store files under (default "$HOME/.countdown")`)
}

func main() {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "countdown")

	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("Missing command:")
		helpMessage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	rest := flag.Args()[1:]

	var err error
	switch cmd {
	case "help":
		helpMessage()
	case "init":
		err = server.InitCmd(countdown.GenInitOptions, logger, *varHome, rest)
	case "start":
		err = server.StartCmd(countdown.GenerateApp, logger, *varHome, rest)
	case "getblock":
		err = server.GetBlockCmd(rest)
	case "retry":
		err = server.RetryCmd(countdown.InlineApp, logger, *varHome, rest)
	case "testgen":
		err = commands.TestGenCmd(countdown.Examples(), rest)
	case "version":
		fmt.Println(Version)
	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		fmt.Printf("Error: %+v\n\n", err)
		helpMessage()
		os.Exit(1)
	}
}
