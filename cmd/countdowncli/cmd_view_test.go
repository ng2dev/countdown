package main

import (
	"bytes"
	"testing"

	countdown "github.com/iov-one/countdown/cmd/countdown/app"
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/x/cash"
)

func TestCmdTransactionView(t *testing.T) {
	tx := &countdown.Tx{
		Sum: &countdown.Tx_CashSendMsg{
			CashSendMsg: &cash.SendMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Memo:     "a memo",
				Ref:      []byte("123"),
			},
		},
	}
	var input bytes.Buffer
	if _, err := writeTx(&input, tx); err != nil {
		t.Fatalf("cannot marshal transaction: %s", err)
	}

	var output bytes.Buffer
	if err := cmdTransactionView(&input, &output, nil); err != nil {
		t.Fatalf("cannot view a transaction: %s", err)
	}

	const want = `{
	"Sum": {
		"CashSendMsg": {
			"metadata": {
				"schema": 1
			},
			"memo": "a memo",
			"ref": "MTIz"
		}
	}
}`
	got := output.String()

	if want != got {
		t.Logf("want: %s", want)
		t.Logf(" got: %s", got)
		t.Fatal("unexpected view result")
	}
}
