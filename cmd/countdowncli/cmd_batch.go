package main

import (
	"errors"
	"flag"
	"fmt"
	"io"

	countdown "github.com/iov-one/countdown/cmd/countdown/app"
	"github.com/iov-one/weave/x/cash"
	"github.com/iov-one/weave/x/multisig"
)

func cmdAsBatch(input io.Reader, output io.Writer, args []string) error {
	fl := flag.NewFlagSet("", flag.ExitOnError)
	fl.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), `
Read any number of transactions from the stdin and extract messages from them.
Create a single batch transaction containing all message. All attributes of the
original transactions (ie signatures) are being dropped.
		`)
		fl.PrintDefaults()
	}
	fl.Parse(args)

	var batch countdown.ExecuteBatchMsg
	for {
		tx, _, err := readTx(input)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		msg, err := tx.GetMsg()
		if err != nil {
			return fmt.Errorf("cannot extract message from the transaction: %s", err)
		}

		// List of all supported batch types can be found in the
		// cmd/countdown/app/codec.proto file.
		//
		// Instead of manually managing this list, use the script from
		// the bottom comment to generate all the cases. Remember to
		// leave the nil and default cases as they are not being
		// generated.
		// You are welcome.
		switch msg := msg.(type) {

		case *cash.SendMsg:
			batch.Messages = append(batch.Messages, countdown.ExecuteBatchMsg_Union{
				Sum: &countdown.ExecuteBatchMsg_Union_CashSendMsg{
					CashSendMsg: msg,
				},
			})
		case *multisig.CreateMsg:
			batch.Messages = append(batch.Messages, countdown.ExecuteBatchMsg_Union{
				Sum: &countdown.ExecuteBatchMsg_Union_MultisigCreateMsg{
					MultisigCreateMsg: msg,
				},
			})
		case *multisig.UpdateMsg:
			batch.Messages = append(batch.Messages, countdown.ExecuteBatchMsg_Union{
				Sum: &countdown.ExecuteBatchMsg_Union_MultisigUpdateMsg{
					MultisigUpdateMsg: msg,
				},
			})
		case nil:
			return errors.New("transaction without a message")
		default:
			return fmt.Errorf("message type not supported: %T", msg)
		}
	}

	batchTx := &countdown.Tx{
		Sum: &countdown.Tx_ExecuteBatchMsg{ExecuteBatchMsg: &batch},
	}
	_, err := writeTx(output, batchTx)
	return err
}

/*
Use this script to generate list of all switch cases for the batch message
building. Make sure that the "protobuf" string contains the most recent
declaration.

#!/bin/bash

# Copy this directly from the ExecuteBatchMsg defined in cmd/countdown/app/codec.proto
protobuf="
cash.SendMsg cash_send_msg = 51;
multisig.CreateMsg multisig_create_msg = 56;
multisig.UpdateMsg multisig_update_msg = 57;
"

while read -r m; do
	if [ "x$m" == "x" ]
	then
		continue
	fi

	tp=`echo $m | cut -d ' ' -f1`
	# Name is not always the same as the type name. Convert it to camel case.
	name=`echo $m | cut -d ' ' -f2 | sed -r 's/(^|_)([a-z])/\U\2/g'`

	echo "	case *$tp:"
	echo "		batch.Messages = append(batch.Messages, countdown.ExecuteBatchMsg_Union{"
	echo "			Sum: &countdown.ExecuteBatchMsg_Union_$name{"
	echo "					$name: msg,"
	echo "				},"
	echo "		})"
done <<< $protobuf
*/
