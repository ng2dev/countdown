package countdown

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/app"
	"github.com/iov-one/weave/errors"

	"github.com/iov-one/weave/store"
	"github.com/iov-one/weave/weavetest"
	"github.com/iov-one/weave/weavetest/assert"
)

const lyrics := []string{
	"(Ten, nine, eight, seven, six, five, four, three, two, one)",
	"We're leaving together",
	"But still it's farewell",
	"And maybe we'll come back",
	"To earth, who can tell?",
	"I guess there is no one to blame",
	"We're leaving ground (leaving ground)",
	"Will things ever be the same again?",
	"It's the final countdown",
	"The final countdown - Oh",
	"We're heading for Venus (Venus)",
	"And still we stand tall",
	"Cause maybe they've seen us (seen us)",
	"And welcome us all, yeah",
	"With so many light years to go",
	"And things to be found (to be found)",
	"I'm sure that we'll all miss her so",
	"It's the final countdown",
	"The final countdown",
	"The final countdown",
	"The final countdown - Oh",
	"It's the final countdown",
	"The final countdown",
	"The final countdown",
	"The final countdown - Oh",
	"It's the final countdown",
}

func TestCreateUser(t *testing.T) {
	cases := map[string]struct {
		msg             weave.Msg
		expected        *User
		wantCheckErrs   map[string]*errors.Error
		wantDeliverErrs map[string]*errors.Error
	}{
		"success": {
			msg: &CreateUserMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Username: "enigma",
			},
			expected: &User{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       weavetest.SequenceID(1),
				Username: "enigma",
			},
			wantCheckErrs: map[string]*errors.Error{
				"Metadata": nil,
				"Username": nil,
			},
			wantDeliverErrs: map[string]*errors.Error{
				"Metadata": nil,
				"Username": nil,
			},
		},
		// TODO add missing metadata test
		"failure missing username": {
			msg: &CreateUserMsg{
				Metadata: &weave.Metadata{Schema: 1},
			},
			expected: nil,
			wantCheckErrs: map[string]*errors.Error{
				"Metadata": nil,
				"Username": errors.ErrModel,
			},
			wantDeliverErrs: map[string]*errors.Error{
				"Metadata": nil,
				"Username": errors.ErrModel,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			auth := &weavetest.Auth{}

			rt := app.NewRouter()

			scheduler := &weavetest.Cron{}
			RegisterRoutes(rt, auth, scheduler)

			kv := store.MemStore()
			bucket := NewUserBucket()

			tx := &weavetest.Tx{Msg: tc.msg}

			ctx := weave.WithBlockTime(context.Background(), time.Now().Round(time.Second))

			if _, err := rt.Check(ctx, kv, tx); err != nil {
				for field, wantErr := range tc.wantCheckErrs {
					assert.FieldError(t, err, field, wantErr)
				}
			}

			res, err := rt.Deliver(ctx, kv, tx)
			if err != nil {
				for field, wantErr := range tc.wantDeliverErrs {
					assert.FieldError(t, err, field, wantErr)
				}
			}

			if tc.expected != nil {
				var stored User
				err := bucket.One(kv, res.Data, &stored)
				assert.Nil(t, err)

				// ensure registeredAt is after test creation time
				registeredAt := stored.RegisteredAt
				weave.InTheFuture(ctx, registeredAt.Time())

				// avoid registered at missing error
				tc.expected.RegisteredAt = registeredAt

				assert.Nil(t, err)
				assert.Equal(t, tc.expected, &stored)
			}
		})
	}
}

func TestCreateCountdown(t *testing.T) {
	owner := weavetest.NewCondition()

	b, err := json.Marshal(lyrics)
	assert.Nil(t, err)

	cases := map[string]struct {
		msg             weave.Msg
		owner           weave.Condition
		expected        *Countdown
		wantCheckErrs   map[string]*errors.Error
		wantDeliverErrs map[string]*errors.Error
	}{
		"success": {
			msg: &CreateCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Title:    "final countdown",
				Lyrics:   b,
			},
			owner: owner,
			expected: &Countdown{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       weavetest.SequenceID(1),
				Owner:    owner.Address(),
				Title:    "final countdown",
				Lyrics:   b,
			},
			wantCheckErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        nil,
				"Owner":     nil,
				"Title":     nil,
				"Lyrics":    nil,
				"Countdown": nil,
			},
			wantDeliverErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        nil,
				"Owner":     nil,
				"Title":     nil,
				"Lyrics":    nil,
				"Countdown": nil,
			},
		},
		// TODO add metadata test
		"failure no signer": {
			msg: &CreateCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Title:    "final countdown",
				Lyrics:   b,
			},
			owner: nil,
			wantCheckErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        nil,
				"Owner":     errors.ErrEmpty,
				"Title":     nil,
				"Lyrics":    nil,
				"Countdown": nil,
			},
			wantDeliverErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        nil,
				"Owner":     errors.ErrEmpty,
				"Title":     nil,
				"Lyrics":    nil,
				"Countdown": nil,
			},
		},
		"failure missing title": {
			msg: &CreateCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Lyrics:   b,
			},
			owner: owner,
			wantCheckErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        nil,
				"Owner":     nil,
				"Title":     errors.ErrModel,
				"Lyrics":    nil,
				"Countdown": nil,
			},
			wantDeliverErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        nil,
				"Owner":     nil,
				"Title":     errors.ErrModel,
				"Lyrics":    nil,
				"Countdown": nil,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			auth := &weavetest.Auth{
				Signer: tc.owner,
			}

			rt := app.NewRouter()

			scheduler := &weavetest.Cron{}
			RegisterRoutes(rt, auth, scheduler)

			kv := store.MemStore()
			bucket := NewCountdownBucket()

			tx := &weavetest.Tx{Msg: tc.msg}

			ctx := weave.WithBlockTime(context.Background(), time.Now().Round(time.Second))

			if _, err := rt.Check(ctx, kv, tx); err != nil {
				for field, wantErr := range tc.wantCheckErrs {
					assert.FieldError(t, err, field, wantErr)
				}
			}

			res, err := rt.Deliver(ctx, kv, tx)
			if err != nil {
				for field, wantErr := range tc.wantDeliverErrs {
					assert.FieldError(t, err, field, wantErr)
				}
			}

			if tc.expected != nil {
				var stored Countdown
				err := bucket.One(kv, res.Data, &stored)
				assert.Nil(t, err)

				// ensure createdAt is after test execution starting time
				createdAt := stored.CreatedAt
				weave.InTheFuture(ctx, createdAt.Time())

				// avoid registered at missing error
				tc.expected.CreatedAt = createdAt

				assert.Nil(t, err)
				assert.Equal(t, tc.expected, &stored)
			}
		})
	}
}

func TestDeleteCountdown(t *testing.T) {
	bob := weavetest.NewCondition()
	signer := weavetest.NewCondition()

	now := weave.AsUnixTime(time.Now())
	future := now.Add(time.Hour)

	b, err := json.Marshal(lyrics)
	assert.Nil(t, err)

	ownedCDID := weavetest.SequenceID(1)
	ownedCD := &Countdown{
		Metadata:  &weave.Metadata{Schema: 1},
		ID:        ownedCDID,
		ID:        weavetest.SequenceID(1),
		Owner:     signer.Address(),
		Title:     "owner's countdown",
		Lyrics:    b,
		CreatedAt: now,
		DeleteAt:  future,
	}

	notOwnedCDID := weavetest.SequenceID(2)
	notOwnedCD := &Countdown{
		Metadata:  &weave.Metadata{Schema: 1},
		ID:        notOwnedCDID,
		ID:        weavetest.SequenceID(2),
		Owner:     bob.Address(),
		Title:     "hacker's countdown",
		Lyrics:    b,
		CreatedAt: now,
		DeleteAt:  future,
	}

	cases := map[string]struct {
		msg             weave.Msg
		signer          weave.Condition
		expected        *Countdown
		wantCheckErrs   map[string]*errors.Error
		wantDeliverErrs map[string]*errors.Error
	}{
		"success": {
			msg: &DeleteCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       ownedCDID,
			},
			signer:   signer,
			expected: ownedCD,
			wantCheckErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"ID":          nil,
				"Owner":       nil,
				"Title":       nil,
				"Lyrics":      nil,
				"Countdown":   nil,
				"CreatedAt":   nil,
				"CompletedAt": nil,
				"DeleteAt":    nil,
			},
			wantDeliverErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"ID":          nil,
				"Owner":       nil,
				"Title":       nil,
				"Lyrics":      nil,
				"Countdown":   nil,
				"CreatedAt":   nil,
				"CompletedAt": nil,
				"DeleteAt":    nil,
			},
		},
		"failure unauthorized": {
			msg: &DeleteCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       notOwnedCDID,
			},
			signer:   signer,
			expected: nil,
			wantCheckErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"ID":          nil,
				"Owner":       nil,
				"Title":       nil,
				"Lyrics":      nil,
				"Countdown":   nil,
				"CreatedAt":   nil,
				"CompletedAt": nil,
				"DeleteAt":    nil,
			},
			wantDeliverErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"ID":          nil,
				"Owner":       nil,
				"Title":       nil,
				"Lyrics":      nil,
				"Countdown":   nil,
				"CreatedAt":   nil,
				"CompletedAt": nil,
				"DeleteAt":    nil,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			auth := &weavetest.Auth{
				Signer: tc.signer,
			}

			// initalize environment
			rt := app.NewRouter()

			scheduler := &weavetest.Cron{}
			RegisterRoutes(rt, auth, scheduler)

			kv := store.MemStore()

			// initalize countdown bucket and save countdowns
			countdownBucket := NewCountdownBucket()
			err := countdownBucket.Put(kv, ownedCD)
			assert.Nil(t, err)

			err = countdownBucket.Put(kv, notOwnedCD)
			assert.Nil(t, err)

			tx := &weavetest.Tx{Msg: tc.msg}

			ctx := weave.WithBlockTime(context.Background(), time.Now().Round(time.Second))

			if _, err := rt.Check(ctx, kv, tx); err != nil {
				for field, wantErr := range tc.wantCheckErrs {
					assert.FieldError(t, err, field, wantErr)
				}
			}

			_, err = rt.Deliver(ctx, kv, tx)
			if err != nil {
				for field, wantErr := range tc.wantDeliverErrs {
					assert.FieldError(t, err, field, wantErr)
				}
			}

			if tc.expected != nil {
				if err := countdownBucket.Has(kv, tc.msg.(*DeleteCountdownMsg).ID); err == nil {
					t.Fatalf("got %+v", err)
				}
			}
		})
	}
}
