package countdown

import (
	"testing"
	"time"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
	"github.com/iov-one/weave/weavetest"
	"github.com/iov-one/weave/weavetest/assert"
)

func TestCountdownUserIDIndexer(t *testing.T) {
	now := weave.AsUnixTime(time.Now())

	userID := weavetest.SequenceID(1)

	cd := &Countdown{
		Metadata:  &weave.Metadata{Schema: 1},
		ID:        weavetest.SequenceID(1),
		Owner:     userID,
		Title:     "Final Countdown",
		CreatedAt: now,
	}

	cases := map[string]struct {
		obj      orm.Object
		expected []byte
		wantErr  *errors.Error
	}{
		"success": {
			obj:      orm.NewSimpleObj(nil, countdown),
			expected: userID,
			wantErr:  nil,
		},
		"failure, obj is nil": {
			obj:      nil,
			expected: nil,
			wantErr:  nil,
		},
		"not countdown": {
			obj:      orm.NewSimpleObj(nil, new(User)),
			expected: nil,
			wantErr:  errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			index, err := countdownUserIDIndexer(tc.obj)

			if !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
			assert.Equal(t, tc.expected, index)
		})
	}
}
