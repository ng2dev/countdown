package countdown

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
	"github.com/iov-one/weave/weavetest"
	"github.com/iov-one/weave/weavetest/assert"
)

func TestValidateUser(t *testing.T) {
	now := weave.AsUnixTime(time.Now())

	cases := map[string]struct {
		model    orm.Model
		wantErrs map[string]*errors.Error
	}{
		"success": {
			model: &User{
				Metadata:     &weave.Metadata{Schema: 1},
				ID:           weavetest.SequenceID(1),
				Username:     "enigma",
				RegisteredAt: now,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":     nil,
				"ID":           nil,
				"Username":     nil,
				"RegisteredAt": nil,
			},
		},
		"failure missing ID": {
			model: &User{
				Metadata:     &weave.Metadata{Schema: 1},
				Username:     "enigma",
				RegisteredAt: now,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":     nil,
				"ID":           errors.ErrEmpty,
				"Username":     nil,
				"RegisteredAt": nil,
			},
		},
		"failure missing username": {
			model: &User{
				Metadata:     &weave.Metadata{Schema: 1},
				ID:           weavetest.SequenceID(1),
				RegisteredAt: now,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":     nil,
				"ID":           nil,
				"Username":     errors.ErrModel,
				"RegisteredAt": nil,
			},
		},
		"failure missing registered at": {
			model: &User{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       weavetest.SequenceID(1),
				Username: "enigma",
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":     nil,
				"ID":           nil,
				"Username":     nil,
				"RegisteredAt": errors.ErrEmpty,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.model.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}

func TestCountdownTask(t *testing.T) {
	cases := map[string]struct {
		model    orm.Model
		wantErrs map[string]*errors.Error
	}{
		"success": {
			model: &CountdownTask{
				Metadata:    &weave.Metadata{Schema: 1},
				ID:          weavetest.SequenceID(1),
				CountdownID: weavetest.SequenceID(1),
				TaskOwner:   weavetest.NewCondition().Address(),
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        nil,
				"ID":        nil,
				"Lyrics":    nil,
				"TaskOwner": nil,
			},
		},
		// TODO add missing metadata test
		"failure missing id": {
			model: &CountdownTask{
				Metadata:  &weave.Metadata{Schema: 1},
				ID:        weavetest.SequenceID(1),
				TaskOwner: weavetest.NewCondition().Address(),
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        errors.ErrEmpty,
				"ID":        nil,
				"Lyrics":    nil,
				"TaskOwner": nil,
			},
		},
		"failure missing countdown id": {
			model: &CountdownTask{
				Metadata:  &weave.Metadata{Schema: 1},
				ID:        weavetest.SequenceID(1),
				TaskOwner: weavetest.NewCondition().Address(),
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        nil,
				"ID":        errors.ErrEmpty,
				"Lyrics":    nil,
				"TaskOwner": nil,
			},
		},
		"failure missing task owner": {
			model: &CountdownTask{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       weavetest.SequenceID(1),
				ID:       weavetest.SequenceID(1),
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":  nil,
				"ID":        nil,
				"ID":        nil,
				"Lyrics":    nil,
				"TaskOwner": errors.ErrEmpty,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.model.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}

func TestValidateCountdown(t *testing.T) {
	now := weave.AsUnixTime(time.Now())
	b, err := json.Marshal(lyrics)
	assert.Nil(t, err)

	cases := map[string]struct {
		model    orm.Model
		wantErrs map[string]*errors.Error
	}{
		"success": {
			model: &Countdown{
				Metadata:  &weave.Metadata{Schema: 1},
				ID:        weavetest.SequenceID(1),
				Owner:     weavetest.NewCondition().Address(),
				Title:     "final countdown",
				Lyrics:    b,
				CreatedAt: now,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"Owner":       nil,
				"Title":       nil,
				"Lyrics":      nil,
				"CreatedAt":   nil,
				"CompletedAt": nil,
			},
		},
		// TODO add missing metadata test
		"failure missing ID": {
			model: &Countdown{
				Metadata:  &weave.Metadata{Schema: 1},
				Owner:     weavetest.NewCondition().Address(),
				Title:     "final countdown",
				Lyrics:    b,
				CreatedAt: now,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          errors.ErrEmpty,
				"Owner":       nil,
				"Title":       nil,
				"Lyrics":      nil,
				"CreatedAt":   nil,
				"CompletedAt": nil,
			},
		},
		"failure missing owner": {
			model: &Countdown{
				Metadata:  &weave.Metadata{Schema: 1},
				ID:        weavetest.SequenceID(1),
				Title:     "final countdown",
				Lyrics:    b,
				CreatedAt: now,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"Owner":       errors.ErrEmpty,
				"Title":       nil,
				"Lyrics":      nil,
				"CreatedAt":   nil,
				"CompletedAt": nil,
			},
		},
		"failure missing title": {
			model: &Countdown{
				Metadata:  &weave.Metadata{Schema: 1},
				ID:        weavetest.SequenceID(1),
				Lyrics:    b,
				Owner:     weavetest.NewCondition().Address(),
				CreatedAt: now,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"Owner":       nil,
				"Title":       errors.ErrModel,
				"Lyrics":      nil,
				"CreatedAt":   nil,
				"CompletedAt": nil,
			},
		},
		"failure missing created at": {
			model: &Countdown{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       weavetest.SequenceID(1),
				Owner:    weavetest.NewCondition().Address(),
				Title:    "final countdown",
				Lyrics:   b,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"Owner":       nil,
				"Title":       nil,
				"Lyrics":      nil,
				"CreatedAt":   errors.ErrEmpty,
				"CompletedAt": nil,
			},
		},
		"failure missing lyrics": {
			model: &Countdown{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       weavetest.SequenceID(1),
				Owner:    weavetest.NewCondition().Address(),
				Title:    "final countdown",
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"Owner":       nil,
				"Title":       nil,
				"Lyrics":      errors.ErrEmpty,
				"CreatedAt":   nil,
				"CompletedAt": nil,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.model.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}
