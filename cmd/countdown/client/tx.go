package client

import (
	countdown "github.com/ng2dev/countdown/cmd/countdown/app"
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/crypto"
	"github.com/iov-one/weave/x/cash"
	"github.com/iov-one/weave/x/sigs"
	"github.com/iov-one/weave/x/validators"
)

// Tx is all the interfaces we need rolled into one
type Tx interface {
	weave.Tx
	sigs.SignedTx
}

// BuildSendTx will create an unsigned tx to move tokens
func BuildSendTx(source, destination weave.Address, amount coin.Coin, memo string) *countdown.Tx {
	return &countdown.Tx{
		Sum: &countdown.Tx_CashSendMsg{
			CashSendMsg: &cash.SendMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				Source:      source,
				Destination: destination,
				Amount:      &amount,
				Memo:        memo,
			},
		},
	}
}

// SignTx modifies the tx in-place, adding signatures
func SignTx(tx *countdown.Tx, signer *crypto.PrivateKey, chainID string, nonce int64) error {
	sig, err := sigs.SignTx(signer, tx, chainID, nonce)
	if err != nil {
		return err
	}
	tx.Signatures = append(tx.Signatures, sig)
	return nil
}

// ParseCountdownTx will load a serialize tx into a format we can read
func ParseCountdownTx(data []byte) (*countdown.Tx, error) {
	var tx countdown.Tx
	err := tx.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// SetValidatorTx will create an unsigned tx to replace current validator set
func SetValidatorTx(u ...weave.ValidatorUpdate) *countdown.Tx {
	return &countdown.Tx{
		Sum: &countdown.Tx_ValidatorsApplyDiffMsg{
			ValidatorsApplyDiffMsg: &validators.ApplyDiffMsg{
				Metadata:         &weave.Metadata{Schema: 1},
				ValidatorUpdates: u,
			},
		},
	}
}
