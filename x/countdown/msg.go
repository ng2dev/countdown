package countdown

import (
	"encoding/json"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/migration"
)

func init() {
	migration.MustRegister(1, &CreateUserMsg{}, migration.NoModification)
	migration.MustRegister(1, &CreateCountdownMsg{}, migration.NoModification)
	migration.MustRegister(1, &DeleteCountdownMsg{}, migration.NoModification)
}

var _ weave.Msg = (*CreateUserMsg)(nil)

// Path returns the routing path for this message.
func (CreateUserMsg) Path() string {
	return "countdown/create_user"
}

// Validate ensures the CreateUserMsg is valid
func (m CreateUserMsg) Validate() error {
	var errs error

	if !validUsername(m.Username) {
		errs = errors.AppendField(errs, "Username", errors.ErrModel)
	}

	return errs
}

var _ weave.Msg = (*CreateCountdownMsg)(nil)

// Path returns the routing path for this message.
func (CreateCountdownMsg) Path() string {
	return "countdown/create_countdown"
}

// Validate ensures the CreateCountdownMsg is valid
func (m CreateCountdownMsg) Validate() error {
	var errs error

	if !validCountdownTitle(m.Title) {
		errs = errors.AppendField(errs, "Title", errors.ErrModel)
	}

	if len(m.Lyrics) == 0 {
		errs = errors.AppendField(errs, "Lyrics", errors.ErrEmpty)
	}

	var lyrics []string

	if err := json.Unmarshal(m.Lyrics, lyrics); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal lyrics for countdown id %s", msg.ID)
	}

	for i, line := range lyrics {
		if !validCountdownLyrics(line) {
			errs = errors.AppendField(errs, "Lyrics line "+string(i), errors.ErrModel)
		}
	}
	return errs
}

var _ weave.Msg = (*DeleteCountdownMsg)(nil)

// Path returns the routing path for this message.
func (DeleteCountdownMsg) Path() string {
	return "countdown/delete_countdown"
}

// Validate ensures DeleteCountdown is valid
func (m DeleteCountdownMsg) Validate() error {
	var errs error

	errs = errors.AppendField(errs, "ID", isGenID(m.ID, false))

	return errs
}
