package countdown

import (
	"time"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/x"
)

const (
	packageName             = "countdown"
	newUserCost       int64 = 1
	newCountdownCost  int64 = 10
	countdownCostUnit int64 = 1000 // first 1000 chars are free then pay 1 per mille
)

// RegisterQuery registers buckets for querying.
func RegisterQuery(qr weave.QueryRouter) {
}

// RegisterRoutes registers handlers for message processing.
func RegisterRoutes(r weave.Registry, auth x.Authenticator, scheduler weave.Scheduler) {
	//r = migration.SchemaMigratingRegistry(packageName, r)
	r.Handle(&CreateUserMsg{}, NewCreateUserHandler(auth))
	r.Handle(&CreateCountdownMsg{}, NewCreateCountdownHandler(auth, scheduler))
	r.Handle(&DeleteCountdownMsg{}, NewDeleteCountdownHandler(auth, scheduler))
}

// RegisterCronRoutes registers routes that are not exposed to
// routers
func RegisterCronRoutes(r weave.Registry, auth x.Authenticator, scheduler weave.Scheduler) {
	r.Handle(&CountdownTask{}, NewCronAddLyricsHandler(auth, scheduler))
}

// ------------------- CreateUserHandler -------------------

// CreateUserHandler will handle CreateUserMsg
type CreateUserHandler struct {
	auth x.Authenticator
	b    *UserBucket
}

var _ weave.Handler = CreateUserHandler{}

// NewCreateUserHandler creates a user message handler
func NewCreateUserHandler(auth x.Authenticator) weave.Handler {
	return CreateUserHandler{
		auth: auth,
		b:    NewUserBucket(),
	}
}

// validate does all common pre-processing between Check and Deliver
func (h CreateUserHandler) validate(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*CreateUserMsg, *User, error) {
	var msg CreateUserMsg

	if err := weave.LoadMsg(tx, &msg); err != nil {
		return nil, nil, errors.Wrap(err, "load msg")
	}

	blockTime, err := weave.BlockTime(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "no block time in header")
	}
	now := weave.AsUnixTime(blockTime)

	user := &User{
		Metadata:     msg.Metadata,
		Username:     msg.Username,
		RegisteredAt: now,
	}

	return &msg, user, nil
}

// Check just verifies it is properly formed and returns
// the cost of executing it.
func (h CreateUserHandler) Check(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*weave.CheckResult, error) {
	_, _, err := h.validate(ctx, store, tx)
	if err != nil {
		return nil, err
	}

	return &weave.CheckResult{GasAllocated: newUserCost}, nil
}

// Deliver creates a custom state and saves if all preconditions are met
func (h CreateUserHandler) Deliver(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*weave.DeliverResult, error) {
	_, user, err := h.validate(ctx, store, tx)
	if err != nil {
		return nil, err
	}

	err = h.b.Put(store, user)
	if err != nil {
		return nil, errors.Wrap(err, "cannot store user")
	}

	// Returns generated user ID as response
	return &weave.DeliverResult{Data: user.ID}, nil
}

// ------------------- CreateCountdownHandler -------------------

// CreateCountdownHandler will handle CreateCountdownMsg
type CreateCountdownHandler struct {
	auth x.Authenticator
	b    *CountdownBucket
	scheduler weave.Scheduler
}

var _ weave.Handler = CreateCountdownHandler{}

// NewCreateCountdownHandler creates a countdown message handler
func NewCreateCountdownandler(auth x.Authenticator, scheduler weave.Scheduler) weave.Handler {
	return CreateCountdownHandler{
		auth: auth,
		b:    NewCountdownBucket(),
		scheduler: scheduler,
	}
}

// validate does all common pre-processing between Check and Deliver
func (h CreateCountdownHandler) validate(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*CreateCountdownMsg, *Countdown, error) {
	var msg CreateCountdownMsg

	if err := weave.LoadMsg(tx, &msg); err != nil {
		return nil, nil, errors.Wrap(err, "load msg")
	}

	blockTime, err := weave.BlockTime(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "no block time in header")
	}
	now := weave.AsUnixTime(blockTime)

	cd := &Countdown{
		Metadata:  msg.Metadata,
		Owner:     x.MainSigner(ctx, h.auth).Address(),
		Title:     msg.Title,
		Lyrics:    msg.Lyrics,
		CreatedAt: now,
	}

	return &msg, cd, nil
}

// Check just verifies it is properly formed and returns
// the cost of executing it.
func (h CreateCountdownHandler) Check(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*weave.CheckResult, error) {
	_, _, err := h.validate(ctx, store, tx)
	if err != nil {
		return nil, err
	}

	return &weave.CheckResult{GasAllocated: newCountdownCost}, nil
}

// Deliver creates an custom state and saves if all preconditions are met
func (h CreateCountdownHandler) Deliver(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*weave.DeliverResult, error) {
	_, cd, err := h.validate(ctx, store, tx)
	if err != nil {
		return nil, err
	}

	if err = h.b.Put(store, cd); err != nil {
		return nil, errors.Wrap(err, "cannot store countdown")
	}

	// schedule first task to be executed for this countdown
	future := weave.AsUnixTime(time.Now().Add(24*time.Hour))
	taskMsg := &CountdownTask{
		Metadata:    msg.Metadata,
		CountdownID: cd.ID,
		TaskOwner: 	 cd.Owner,
	}

	if _, err := h.scheduler.Schedule(store, future, nil, taskMsg); err != nil {
		return nil, errors.Wrap(err, "could not schedule task")
	}

	// Returns generated countdown ID as response
	return &weave.DeliverResult{Data: cd.ID}, nil
}

// ------------------- DeleteCountdownHandler -------------------

// DeleteCountdownHandler will handle DeleteCountdownMsg
type DeleteCountdownHandler struct {
	auth x.Authenticator
	b    *CountdownBucket
	scheduler weave.Scheduler
}

var _ weave.Handler = DeleteCountdownHandler{}

// DeleteCountdownHandler creates a countdown message handler
func NewDeleteCountdownHandler(auth x.Authenticator, scheduler weave.Scheduler) weave.Handler {
	return DeleteCountdownHandler{
		auth: auth,
		b:    NewCountdownBucket(),
		scheduler: scheduler,
	}
}

// validate does all common pre-processing between Check and Deliver
func (h DeleteCountdownHandler) validate(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*DeleteCountdownMsg, *Countdown, error) {
	var msg DeleteCountdownMsg

	if err := weave.LoadMsg(tx, &msg); err != nil {
		return nil, nil, errors.Wrap(err, "load msg")
	}

	var cd Countdown
	if err := h.b.One(store, msg.ID, &cd); err != nil {
		return nil, nil, errors.Wrapf(err, "cannot retrieve countdown with ID %s", msg.ID)
	}

	signer := x.MainSigner(ctx, h.auth).Address()
	if !cd.Owner.Equals(signer) {
		return nil, nil, errors.Wrapf(errors.ErrUnauthorized, "signer %s is unauthorized to delete countdown with ID %s", signer, cd.ID)
	}

	return &msg, &cd, nil
}

// Check just verifies it is properly formed and returns
// the cost of executing it.
func (h DeleteCountdownHandler) Check(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*weave.CheckResult, error) {
	_, _, err := h.validate(ctx, store, tx)
	if err != nil {
		return nil, err
	}

	// Deleting is free of charge
	return &weave.CheckResult{}, nil
}

// Deliver creates an custom state and saves if all preconditions are met
func (h DeleteCountdownHandler) Deliver(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*weave.DeliverResult, error) {
	_, cd, err := h.validate(ctx, store, tx)
	if err != nil {
		return nil, err
	}

	if err := h.b.Delete(store, cd.ID); err != nil {
		return nil, errors.Wrapf(err, "cannot delete countdown with ID %s", cd.ID)
	}

	return &weave.DeliverResult{}, nil
}

// ------------------- CronAddLyricsHandler -------------------

// CronAddLyricsHandler will handle scheduled CountdownTask
type CronAddLyricsHandler struct {
	auth x.Authenticator
	b    *CountdownBucket
	scheduler weave.Scheduler
}

var _ weave.Handler = CronAddCountdownHandler{}

// NewCronAddCountdownHandler creates a countdown message handler
func NewCronAddLyricsHandler(auth x.Authenticator, scheduler weave.Scheduler) weave.Handler {
	return CronAddLyricsHandler{
		auth: auth,
		b:    NewCountdownBucket(),
		scheduler: scheduler,
	}
}

// validate does all common pre-processing between Check and Deliver
func (h CronAddLyricsHandler) validate(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*CountdownTask, error) {
	var msg CountdownTask

	if err := weave.LoadMsg(tx, &msg); err != nil {
		return nil, errors.Wrap(err, "load msg")
	}

	var cd Countdown
	if err := h.b.One(store, msg.ID, &countdown); err != nil {
		return nil, nil, errors.Wrapf(err, "cannot retrieve countdown with id %s", msg.ID)
	}

	return &msg, nil
}

// Check just verifies it is properly formed and returns
// the cost of executing it.
func (h CronAddLyricsHandler) Check(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*weave.CheckResult, error) {
	_, err := h.validate(ctx, store, tx)
	if err != nil {
		return nil, err
	}

	return &weave.CheckResult{}, nil
}

// Deliver stages a scheduled addition of lyrics to the countdown if all preconditions are met
func (h CronAddLyricsHandler) Deliver(ctx weave.Context, store weave.KVStore, tx weave.Tx) (*weave.DeliverResult, error) {
	msg, err := h.validate(ctx, store, tx)
	if err != nil {
		return nil, err
	}

	var cd Countdown
	if err := h.b.One(store, msg.ID, &countdown); err != nil {
		return nil, nil, errors.Wrapf(err, "cannot retrieve countdowns %s", msg.ID)
	}

	if len(cd.Countdown) < len(cd.Lyrics) {
		var lyrics []string
		var countdown []string

		if err := json.Unmarshal(cd.Lyrics, lyrics); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal lyrics for countdown id %s", msg.ID)
		}

		if err := json.Unmarshal(cd.Countdown, countdown); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal countdown lyrics for countdown id %s", msg.ID)
		}
		// append a new line of lyrics to the countdown
		countdown = append(countdown, lyrics[len(countdown)])

		cd.Countdown, err := json.Marshal(countdown)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal added lyrics for countdown id %s", msg.ID)
		}

		// schedule next task to be executed
		future := weave.AsUnixTime(time.Now().Add(24*time.Hour))
		taskMsg := &CountdownTask{
			Metadata:  msg.Metadata,
			CountdownID: cd.ID,
			TaskOwner: cd.Owner,
		}
	
		if _, err := h.scheduler.Schedule(store, future, nil, taskMsg); err != nil {
			return nil, errors.Wrap(err, "could not schedule next task")
		}

	} else {
		// the countdown has reached its final line and is marked completed
		cd.CompletedAt = weave.AsUnixTime(time.Now())
	}

	if err := h.b.Put(store, cd); err != nil {
		return nil, errors.Wrapf(err, "cannot add lyrics to countdown with ID %s", msg.ID)
	}

	return &weave.DeliverResult{}, nil
}
