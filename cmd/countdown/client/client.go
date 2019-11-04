package client

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/app"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/x/sigs"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// BroadcastTxSyncDefaultTimeOut timeout for sync tx broadcasting
const BroadcastTxSyncDefaultTimeOut = 15 * time.Second

var QueryNewBlockHeader = tmtypes.EventQueryNewBlockHeader

// Client is an interface to interact with weave apps
type Client interface {
	// TendermintClient returns the underlying tendermint client
	TendermintClient() client.Client
	// GetUser will return nonce and public key registered
	// for a given address if it was ever used.
	GetUser(addr weave.Address) (*UserResponse, error)
	// GetWallet will return a wallet given an address
	GetWallet(addr weave.Address) (*WalletResponse, error)
	// BroadcastTx serializes a signed transaction and writes to the
	// blockchain. It returns when the tx is committed to the blockchain.
	BroadcastTx(tx weave.Tx) BroadcastTxResponse
	// BroadcastTxAsync can be run in a goroutine and will output the
	// result or error to the given channel.
	BroadcastTxAsync(tx weave.Tx, out chan<- BroadcastTxResponse)
	// BroadcastTxSync brodcasts transactions synchronously
	BroadcastTxSync(tx weave.Tx, timeout time.Duration) BroadcastTxResponse
	// AbciQuery calls abci query on tendermint rpc.
	AbciQuery(path string, data []byte) (AbciResponse, error)
	// NextNonce queries the blockchain for the next nonce
	NextNonce(client Client, addr weave.Address) (int64, error)
}

// CountdownClient is a tendermint client wrapped to provide
// simple access to the data structures used in the countdown module.
type CountdownClient struct {
	conn client.Client
	// subscriber is a unique identifier for subscriptions
	subscriber string
}

// NewClient wraps a CountdownClient around an existing
// tendermint client connection.
func NewClient(conn client.Client) *CountdownClient {
	return &CountdownClient{
		conn:       conn,
		subscriber: "tools-client",
	}
}

// TendermintClient returns underlying tendermint client
func (cc *CountdownClient) TendermintClient() client.Client {
	return cc.conn
}

//************ generic (weave) functionality *************//

// Status will return the raw status from the node
func (cc *CountdownClient) Status() (*ctypes.ResultStatus, error) {
	return cc.conn.Status()
}

// Genesis will return the genesis directly from the node
func (cc *CountdownClient) Genesis() (*tmtypes.GenesisDoc, error) {
	gen, err := cc.conn.Genesis()
	if err != nil {
		return nil, err
	}
	return gen.Genesis, nil
}

// ChainID will parse out the chainID from the status result
func (cc *CountdownClient) ChainID() (string, error) {
	gen, err := cc.Genesis()
	if err != nil {
		return "", err
	}
	return gen.ChainID, nil
}

// Height will parse out the Height from the status result
func (cc *CountdownClient) Height() (int64, error) {
	status, err := cc.conn.Status()
	if err != nil {
		return -1, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

// AbciResponse contains a query result:
// a (possibly empty) list of key-value pairs, and the height
// at which it queried
type AbciResponse struct {
	// a list of key/value pairs
	Models []weave.Model
	Height int64
}

// AbciQuery calls abci query on tendermint rpc,
// verifies if it is an error or empty, and if there is
// data pulls out the ResultSets from keys and values into
// a useful AbciResponse struct
func (cc *CountdownClient) AbciQuery(path string, data []byte) (AbciResponse, error) {
	var out AbciResponse

	q, err := cc.conn.ABCIQuery(path, data)
	if err != nil {
		return out, err
	}
	resp := q.Response
	if resp.IsErr() {
		return out, errors.ABCIError(resp.Code, resp.Log)
	}
	out.Height = resp.Height

	if len(resp.Key) == 0 {
		return out, nil
	}

	// assume there is data, parse the result sets
	var keys, vals app.ResultSet
	err = keys.Unmarshal(resp.Key)
	if err != nil {
		return out, err
	}
	err = vals.Unmarshal(resp.Value)
	if err != nil {
		return out, err
	}

	out.Models, err = app.JoinResults(&keys, &vals)
	return out, err
}

// TxSearch searches transactions using underlying tendermint client
func (cc *CountdownClient) TxSearch(query string, prove bool, page, perPage int) (*ctypes.ResultTxSearch, error) {
	return cc.conn.TxSearch(query, prove, page, perPage)
}

// BroadcastTxResponse is the result of submitting a transaction.
type BroadcastTxResponse struct {
	Error    error                           // not-nil if there was an error sending
	Response *ctypes.ResultBroadcastTxCommit // not-nil if we got response from node
}

// IsError returns the error for failure if it failed,
// or null if it succeeded
func (b BroadcastTxResponse) IsError() error {
	if b.Error != nil {
		return b.Error
	}
	if b.Response.CheckTx.IsErr() {
		ctx := b.Response.CheckTx
		return errors.Wrap(errors.ABCIError(ctx.Code, ctx.Log), "CheckTx error")
	}
	if b.Response.DeliverTx.IsErr() {
		dtx := b.Response.DeliverTx
		return errors.Wrap(errors.ABCIError(dtx.Code, dtx.Log), "DeliverTx error")
	}
	return nil
}

// BroadcastTx serializes a signed transaction and writes to the
// blockchain. It returns when the tx is committed to the
// blockchain.
//
// If you want high-performance, parallel sending, use BroadcastTxAsync
func (cc *CountdownClient) BroadcastTx(tx weave.Tx) BroadcastTxResponse {
	out := make(chan BroadcastTxResponse, 1)
	defer close(out)
	go cc.BroadcastTxAsync(tx, out)
	res := <-out
	return res
}

// BroadcastTxSync brodcasts transactions synchronously
func (cc *CountdownClient) BroadcastTxSync(tx weave.Tx, timeout time.Duration) BroadcastTxResponse {
	data, err := tx.Marshal()
	if err != nil {
		return BroadcastTxResponse{Error: err}
	}

	res, err := cc.conn.BroadcastTxSync(data)
	if err != nil {
		return BroadcastTxResponse{Error: err}
	}
	if res.Code != 0 {
		err = errors.Wrap(errors.ABCIError(res.Code, res.Log), "CheckTx error")
		return BroadcastTxResponse{Error: err}
	}

	// and wait for confirmation
	evt, err := cc.WaitForTxEvent(data, tmtypes.EventTx, timeout)
	if err != nil {
		return BroadcastTxResponse{Error: err}
	}

	txe, ok := evt.(tmtypes.EventDataTx)
	if !ok {
		if err != nil {
			err = errors.Wrap(err, "WaitForOneEvent did not return an EventDataTx object")
			return BroadcastTxResponse{Error: err}
		}
	}

	return BroadcastTxResponse{
		Response: &ctypes.ResultBroadcastTxCommit{
			DeliverTx: txe.Result,
			Height:    txe.Height,
			Hash:      txe.Tx.Hash(),
		},
	}
}

// WaitForTxEvent listens for and particular event type of evtTyp to be fired
func (cc *CountdownClient) WaitForTxEvent(tx tmtypes.Tx, evtTyp string, timeout time.Duration) (tmtypes.TMEventData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	query := tmtypes.EventQueryTxFor(tx)

	uuid := hex.EncodeToString(append(tx.Hash(), cmn.RandBytes(2)...))
	evts, err := cc.conn.Subscribe(ctx, uuid, query.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to subscribe")
	}

	// make sure to unregister after the test is over
	defer cc.conn.UnsubscribeAll(ctx, uuid)

	select {
	case evt := <-evts:
		return evt.Data.(tmtypes.TMEventData), nil
	case <-ctx.Done():
		return nil, errors.Wrap(errors.ErrTimeout, "waiting for event timed out")
	}
}

// BroadcastTxAsync can be run in a goroutine and will output
// the result or error to the given channel.
// Useful if you want to send many tx in parallel
func (cc *CountdownClient) BroadcastTxAsync(tx weave.Tx, out chan<- BroadcastTxResponse) {
	data, err := tx.Marshal()
	if err != nil {
		out <- BroadcastTxResponse{Error: err}
		return
	}

	// TODO: make this async, maybe adjust return value
	res, err := cc.conn.BroadcastTxCommit(data)
	msg := BroadcastTxResponse{
		Error:    err,
		Response: res,
	}
	out <- msg
}

// SubscribeHeaders queries for headers and starts a goroutine
// to typecase the events into Headers. Returns a cancel
// function. If you don't want the automatic goroutine, use
// Subscribe(QueryNewBlockHeader, out)
func (cc *CountdownClient) SubscribeHeaders(out chan<- *tmtypes.Header) (func(), error) {
	query := tmtypes.EventQueryNewBlockHeader
	pipe, cancel, err := cc.Subscribe(query)
	if err != nil {
		return nil, err
	}
	go func() {
		for msg := range pipe {
			evt, ok := msg.Data.(tmtypes.EventDataNewBlockHeader)
			if !ok {
				// TODO: something else?
				panic("Unexpected event type")
			}
			out <- &evt.Header
		}
		close(out)
	}()
	return cancel, nil
}

// Subscribe will take an arbitrary query and push all events to
// the given channel. If there is no error,
// returns a cancel function that can be called to cancel
// the subscription
func (cc *CountdownClient) Subscribe(query tmpubsub.Query) (<-chan ctypes.ResultEvent, func(), error) {
	ctx := context.Background()
	out, err := cc.conn.Subscribe(ctx, cc.subscriber, query.String())
	if err != nil {
		return out, nil, err
	}
	cancel := func() {
		cc.conn.Unsubscribe(ctx, cc.subscriber, query.String())
	}
	return out, cancel, nil
}

// UnsubscribeAll cancels all subscriptions
func (cc *CountdownClient) UnsubscribeAll() error {
	ctx := context.Background()
	return cc.conn.UnsubscribeAll(ctx, cc.subscriber)
}

// GetWallet will return a wallet given an address
// If non wallet is present, it will return (nil, nil)
// Error codes are used when the query failed on the server
func (cc *CountdownClient) GetWallet(addr weave.Address) (*WalletResponse, error) {
	// make sure we send a valid address to the server
	err := addr.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "invalid address")
	}

	resp, err := cc.AbciQuery("/wallets", addr)
	if err != nil {
		return nil, err
	}
	if len(resp.Models) == 0 { // empty list or nil
		return nil, errors.Wrap(errors.ErrNotFound, "model not found")
	}
	// assume only one result
	model := resp.Models[0]
	// make sure the return value is expected
	acct := walletKeyToAddr(model.Key)
	if !addr.Equals(acct) {
		return nil, errors.Wrapf(ErrNoMatch, "queried %s, returned %s", addr, acct)
	}
	out := WalletResponse{
		Address: acct,
		Height:  resp.Height,
	}

	// parse the value as wallet bytes
	err = out.Wallet.Unmarshal(model.Value)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// key is the address prefixed with "wallet:"
func walletKeyToAddr(key []byte) weave.Address {
	return key[5:]
}

// UserResponse is a response on a query for a User
type UserResponse struct {
	Address  weave.Address
	UserData sigs.UserData
	Height   int64
}

// GetUser will return nonce and public key registered
// for a given address if it was ever used.
// If it returns (nil, nil), then this address never signed
// a transaction before (and can use nonce = 0)
func (cc *CountdownClient) GetUser(addr weave.Address) (*UserResponse, error) {
	// make sure we send a valid address to the server
	err := addr.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "invalid address")
	}

	resp, err := cc.AbciQuery("/auth", addr)
	if err != nil {
		return nil, err
	}
	if len(resp.Models) == 0 { // empty list or nil
		return nil, nil
	}
	// assume only one result
	model := resp.Models[0]

	// make sure the return value is expected
	acct := userKeyToAddr(model.Key)
	if !addr.Equals(acct) {
		return nil, errors.Wrapf(ErrNoMatch, "queried %s, returned %s", addr, acct)
	}
	out := UserResponse{
		Address: acct,
		Height:  resp.Height,
	}

	// parse the value as wallet bytes
	err = out.UserData.Unmarshal(model.Value)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// key is the address prefixed with "sigs:"
func userKeyToAddr(key []byte) weave.Address {
	return key[5:]
}

// NextNonce queries the blockchain for the next nonce
// returns 0 if the address never used
func (cc *CountdownClient) NextNonce(addr weave.Address) (int64, error) {
	user, err := cc.GetUser(addr)
	if err != nil {
		return 0, err
	}
	if user != nil {
		return user.UserData.Sequence, nil
	}
	// new account starts at 0
	return 0, nil
}
