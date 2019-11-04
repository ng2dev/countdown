package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/iov-one/countdown/cmd/countdown/client"
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/x/batch"
)

func cmdSubmitTransaction(input io.Reader, output io.Writer, args []string) error {
	fl := flag.NewFlagSet("", flag.ExitOnError)
	fl.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), `
Read binary serialized transaction from standard input and submit it.

For certain transactions response is written out. If a batch transaction was
submitted, multiple responses can be printed out, one for each message
submitted as part of the batch.

Make sure to collect enough signatures before submitting the transaction.
`)
		fl.PrintDefaults()
	}
	var (
		tmAddrFl = fl.String("tm", env("COUNTDOWNCLI_TM_ADDR", "https://COUNTDOWN.NETWORK.iov.one:443"),
			"Tendermint node address. Use proper NETWORK name. You can use COUNTDOWNCLI_TM_ADDR environment variable to set it.")
	)
	fl.Parse(args)

	tx, _, err := readTx(input)
	if err != nil {
		return fmt.Errorf("cannot read transaction from input: %s", err)
	}

	cl := client.NewClient(client.NewHTTPConnection(*tmAddrFl))

	resp := cl.BroadcastTx(tx)
	if resp.IsError(); err != nil {
		return fmt.Errorf("cannot broadcast transaction: %s", err)
	}

	responses, err := extractResponse(tx, resp.Response.DeliverTx.Data, formatters)
	if err != nil {
		return fmt.Errorf("cannot extract response: %s", err)
	}
	for _, r := range responses {
		fmt.Fprintln(output, r)
	}
	return nil
}

// extractResponses parse given raw response data bytes according to what is
// expected considering the submitted transaction. It returns a human readable
// representation of given response. It can return no data (and no error) if
// response does not contain anythink worth showing to the user or response is
// not supported.
func extractResponse(tx weave.Tx, respData []byte, fmts map[string]func([]byte) (string, error)) ([]string, error) {
	var (
		msgs          []weave.Msg
		responsesData [][]byte
	)
	msg, err := tx.GetMsg()
	if err != nil {
		return nil, fmt.Errorf("cannot extract message from transaction: %s", err)
	}

	// check if msg is a batch message
	if b, ok := msg.(batch.Msg); ok {
		bmsgs, err := b.MsgList()
		if err != nil {
			return nil, fmt.Errorf("cannot extract messages from a batch message trasnaction: %s", err)
		}
		msgs = append(msgs, bmsgs...)
		var arr batch.ByteArrayList
		if err := arr.Unmarshal(respData); err != nil {
			return nil, fmt.Errorf("cannot unmarshal batch message transaction response: %s", err)
		}
		responsesData = arr.Elements
	} else {
		// if not batch message map to weave.Msg
		msgs = []weave.Msg{msg}
		responsesData = [][]byte{respData}
	}

	var responses []string
	for i, msg := range msgs {
		// check if there is a specified format
		format, ok := fmts[msg.Path()]
		if !ok {
			// If no formatter is registered, we do not print the result.
			continue
		}
		pretty, err := format(responsesData[i])
		if err != nil {
			return nil, fmt.Errorf("cannot format #%d result data %x: %s", i, responsesData[i], err)
		}

		responses = append(responses, pretty)
	}
	return responses, nil
}

// formatters contains a mapping of a message path to response parser. Response
// parse function accepts a raw bytes of serialized response and must return a
// human representation of that data.
//
// Do not register a message if you want response returned after its submission
// to be ignored (not printed to the user).
var formatters = map[string]func([]byte) (string, error){
	// add desired format as :
	// gov.CreateTextResolutionMsg{}.Path(): fmtSequence,
	// escrow.CreateMsg{}.Path():            fmtSequence,
}

func fmtSequence(raw []byte) (string, error) {
	n, err := fromSequence(raw)
	if err != nil {
		return "", fmt.Errorf("cannot parse sequence: %s", err)
	}
	return fmt.Sprint(n), nil
}
