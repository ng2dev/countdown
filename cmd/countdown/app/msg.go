package countdown

import (
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/x/batch"
)

// Boiler-plate needed to bridge the ExecuteBatchMsg protobuf type into something usable by the batch extension
var _ batch.Msg = (*ExecuteBatchMsg)(nil)

// Path returns path of execute message
func (*ExecuteBatchMsg) Path() string {
	return batch.PathExecuteBatchMsg
}

// Validate validates execute message
func (msg *ExecuteBatchMsg) Validate() error {
	return batch.Validate(msg)
}

// MsgList decode msg.Messages to weave.Msg array
func (msg *ExecuteBatchMsg) MsgList() ([]weave.Msg, error) {
	var err error
	messages := make([]weave.Msg, len(msg.Messages))
	for i, m := range msg.Messages {
		messages[i], err = weave.ExtractMsgFromSum(m.GetSum())
		if err != nil {
			return nil, err
		}
	}
	return messages, nil
}
