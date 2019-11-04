package countdown

import (
	"encoding/json"
	"testing"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/weavetest"
	"github.com/iov-one/weave/weavetest/assert"
)

func TestValidateCreateUserMsg(t *testing.T) {
	cases := map[string]struct {
		msg      weave.Msg
		wantErrs map[string]*errors.Error
	}{
		"success": {
			msg: &CreateUserMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Username: "enigma",
			},
			wantErrs: map[string]*errors.Error{
				"Metadata": nil,
				"Username": nil,
			},
		},
		"failure missing username": {
			msg: &CreateUserMsg{
				Metadata: &weave.Metadata{Schema: 1},
			},
			wantErrs: map[string]*errors.Error{
				"Metadata": nil,
				"Username": errors.ErrModel,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.msg.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}

func TestValidateCreateCountdownMsg(t *testing.T) {
	b, err := json.Marshal(lyrics)
	assert.Nil(t, err)

	cases := map[string]struct {
		msg      weave.Msg
		wantErrs map[string]*errors.Error
	}{
		"success": {
			msg: &CreateCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Title:    "final countdown",
				Lyrics:   b,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata": nil,
				"Title":    nil,
				"Lyrics":   nil,
			},
		},
		"failure missing title": {
			msg: &CreateCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Lyrics:   b,
			},
			wantErrs: map[string]*errors.Error{
				"Metadata": nil,
				"Title":    errors.ErrModel,
				"Lyrics":   nil,
			},
		},
		"failure missing lyrics": {
			msg: &CreateCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Title:    "final countdown",
			},
			wantErrs: map[string]*errors.Error{
				"Metadata": nil,
				"Title":    nil,
				"Lyrics":   errors.ErrEmpty,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.msg.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}

func TestValidateDeleteCountdown(t *testing.T) {
	cases := map[string]struct {
		msg      weave.Msg
		wantErrs map[string]*errors.Error
	}{
		"success": {
			msg: &DeleteCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       weavetest.SequenceID(1),
			},
			wantErrs: map[string]*errors.Error{
				"Metadata": nil,
				"ID":       nil,
			},
		},
		// add missing metadata test
		"failure missing article id": {
			msg: &DeleteCountdownMsg{
				Metadata: &weave.Metadata{Schema: 1},
			},
			wantErrs: map[string]*errors.Error{
				"Metadata": nil,
				"ID":       errors.ErrEmpty,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.msg.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}

func TestCountdownTask(t *testing.T) {
	cases := map[string]struct {
		msg      weave.Msg
		wantErrs map[string]*errors.Error
	}{
		"success": {
			msg: &CountdownTaskMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				ID:          weavetest.SequenceID(1),
				CountdownID: weavetest.SequenceID(1),
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          nil,
				"CountdownID": nil,
			},
		},
		// add missing metadata test
		"failure missing task id": {
			msg: &CountdownTaskMsg{
				Metadata: &weave.Metadata{Schema: 1},
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          errors.ErrEmpty,
				"CountdownID": nil,
			},
		},
		"failure invalid task id": {
			msg: &CountdownTaskMsg{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       []byte{0, 0},
			},
			wantErrs: map[string]*errors.Error{
				"Metadata":    nil,
				"ID":          errors.ErrInput,
				"CountdownID": nil,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.msg.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}
