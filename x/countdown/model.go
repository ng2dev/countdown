package countdown

import (
	"encoding/json"
	"regexp"

	"github.com/iov-one/blog-tutorial/morm"

	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

var _ morm.Model = (*User)(nil)

// SetID is a minimal implementation, useful when the ID is a separate protobuf field
func (m *User) SetID(id []byte) error {
	m.ID = id
	return nil
}

// Copy produces a new copy to fulfill the Model interface
func (m *User) Copy() orm.CloneableData {
	return &User{
		Metadata:     m.Metadata.Copy(),
		ID:           copyBytes(m.ID),
		Username:     m.Username,
		RegisteredAt: m.RegisteredAt,
	}
}

var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_.-]{4,16}$`).MatchString

// Validate validates user's fields
func (m *User) Validate() error {
	var errs error

	errs = errors.AppendField(errs, "ID", isGenID(m.ID, false))

	if !validUsername(m.Username) {
		errs = errors.AppendField(errs, "Username", errors.ErrModel)
	}

	if err := m.RegisteredAt.Validate(); err != nil {
		errs = errors.AppendField(errs, "RegisteredAt", m.RegisteredAt.Validate())
	} else if m.RegisteredAt == 0 {
		errs = errors.AppendField(errs, "RegisteredAt", errors.ErrEmpty)
	}

	return errs
}

var _ morm.Model = (*Countdown)(nil)

// SetID is a minimal implementation, useful when the ID is a separate protobuf field
func (m *Countdown) SetID(id []byte) error {
	m.ID = id
	return nil
}

// Copy produces a new copy to fulfill the Model interface
func (m *Countdown) Copy() orm.CloneableData {
	return &Countdown{
		Metadata:  m.Metadata.Copy(),
		ID:        copyBytes(m.ID),
		Owner:     m.Owner.Clone(),
		Title:     m.Title,
		Lyrics:    copyBytes(m.Lyrics),
		Countdown: copyBytes(m.Countdown),
		CreatedAt: m.CreatedAt,
	}
}

var validCountdownTitle = regexp.MustCompile(`^[a-zA-Z0-9$@$!%*?&#'^;-_. +]{4,32}$`).MatchString
var validCountdownLyrics = regexp.MustCompile(`^[a-zA-Z0-9$@$!%*?&#'^;-_. +]{4,1000}$`).MatchString

// Validate validates countdown's fields
func (m *Countdown) Validate() error {
	var errs error

	errs = errors.AppendField(errs, "ID", isGenID(m.ID, false))
	errs = errors.AppendField(errs, "Owner", m.Owner.Validate())

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

	if err := m.CreatedAt.Validate(); err != nil {
		errs = errors.AppendField(errs, "CreatedAt", err)
	} else if m.CreatedAt == 0 {
		errs = errors.AppendField(errs, "CreatedAt", errors.ErrEmpty)
	}

	return errs
}

var _ morm.Model = (*CountdownTask)(nil)

// SetID is a minimal implementation, useful when the ID is a separate protobuf field
func (m *CountdownTask) SetID(id []byte) error {
	m.ID = id
	return nil
}

// Copy produces a new copy to fulfill the Model interface
func (m *CountdownTask) Copy() orm.CloneableData {
	return &CountdownTask{
		Metadata:  m.Metadata.Copy(),
		ID:        copyBytes(m.ID),
		TaskOwner: m.TaskOwner.Clone(),
	}
}

// Validate validates user's fields
func (m *CountdownTask) Validate() error {
	var errs error

	errs = errors.AppendField(errs, "ID", isGenID(m.ID, false))
	errs = errors.AppendField(errs, "TaskOwner", m.TaskOwner.Validate())

	return errs
}

func copyBytes(in []byte) []byte {
	if in == nil {
		return nil
	}
	cpy := make([]byte, len(in))
	copy(cpy, in)
	return cpy
}

// isGenID ensures that the ID is 8 byte input.
// if allowEmpty is set, we also allow empty
// TODO change with validateSequence when weave 0.22.0 is released
func isGenID(id []byte, allowEmpty bool) error {
	if len(id) == 0 {
		if allowEmpty {
			return nil
		}
		return errors.Wrap(errors.ErrEmpty, "missing id")
	}
	if len(id) != 8 {
		return errors.Wrap(errors.ErrInput, "id must be 8 bytes")
	}
	return nil
}
