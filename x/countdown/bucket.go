package countdown

import (
	"github.com/iov-one/blog-tutorial/morm"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

type UserBucket struct {
	morm.ModelBucket
}

// NewUserBucket returns a new user bucket
func NewUserBucket() *UserBucket {
	return &UserBucket{
		morm.NewModelBucket("user", &User{}),
	}
}

type CountdownBucket struct {
	morm.ModelBucket
}

// NewCountdownBucket returns a new countdown bucket
func NewCountdownBucket() *CountdownBucket {
	return &CountdownBucket{
		morm.NewModelBucket("countdown", &Countdown{},
			morm.WithIndex("user", countdownUserIDIndexer, false)),
	}
}

// countdownUserIDIndexer enables querying countdowns by user ids
func countdownUserIDIndexer(obj orm.Object) ([]byte, error) {
	if obj == nil || obj.Value() == nil {
		return nil, nil
	}
	cd, ok := obj.Value().(*Countdown)
	if !ok {
		return nil, errors.Wrapf(errors.ErrState, "expected countdown, got %T", obj.Value())
	}
	return cd.Owner, nil
}

type CountdownTaskBucket struct {
	morm.ModelBucket
}

// NewCountdownTaskBucket returns a new lyrics task bucket
func NewCountdownTaskBucket() *CountdownTaskBucket {
	return &CountdownTaskBucket{
		morm.NewModelBucket("tasks", &CountdownTaskBucket{}),
	}
}
