package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
)

func cmdTransactionView(input io.Reader, output io.Writer, args []string) error {
	fl := flag.NewFlagSet("", flag.ExitOnError)
	fl.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), `
Decode and display transaction summary. This command is helpful when reciving a
binary representation of a transaction. Before signing you should check what
kind of operation are you authorizing.
`)
		fl.PrintDefaults()
	}
	fl.Parse(args)

	for {
		var buf bytes.Buffer
		tx, _, err := readTx(io.TeeReader(input, &buf))
		if err == nil {
			// Protobuf compiler is exposing all attributes as JSON as
			// well. This will produce a beautiful summary.
			pretty, err := json.MarshalIndent(tx, "", "\t")
			if err != nil {
				return fmt.Errorf("cannot JSON serialize: %s", err)
			}
			_, _ = output.Write(pretty)

			// if you want to print extra info from message you can extract
			// and print additionally.
			// _ = printProposalMsg(output, tx)
			return nil
		}
		if err == io.EOF {
			return nil
		}

		// if msg is not a tx you can try as non TX payload
		// such as:
		// 	msg, err := readProposalPayloadMsg(&buf)
		// 	if err != nil {
		//		return err
		// 	}
		// 	pretty, err := json.MarshalIndent(msg, "", "\t")
		// 	if err != nil {
		// 	 	return fmt.Errorf("cannot JSON serialize: %s", err)
		// 	}
		// 	_, _ = output.Write(pretty)
		// 	return nil
	}
}
