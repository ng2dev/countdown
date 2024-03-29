package countdown

import (
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/x/sigs"
)

// TxDecoder creates a Tx and unmarshals bytes into it
func TxDecoder(bz []byte) (weave.Tx, error) {
	tx := new(Tx)
	err := tx.Unmarshal(bz)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// make sure tx fulfills all interfaces
var _ weave.Tx = (*Tx)(nil)
var _ sigs.SignedTx = (*Tx)(nil)

// GetMsg returns a single message instance that is represented by this transaction.
func (tx *Tx) GetMsg() (weave.Msg, error) {
	return weave.ExtractMsgFromSum(tx.GetSum())
}

// GetSignBytes returns the bytes to sign...
func (tx *Tx) GetSignBytes() ([]byte, error) {
	// temporarily unset the signatures, as the sign bytes
	// should only come from the data itself, not previous signatures
	sigs := tx.Signatures
	tx.Signatures = nil

	bz, err := tx.Marshal()

	// reset the signatures after calculating the bytes
	tx.Signatures = sigs
	return bz, err
}
