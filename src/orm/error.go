package orm

import (
	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
	"gorm.io/gorm"
)

const (
	ErrNotFound = "not_found"
)

func wrapListResult[T any](m []*T, err error) ([]*T, *errors.Error) {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Wrap(err, ErrNotFound)
		}
		return nil, errors.Wrap(err, "")
	}

	return m, nil
}

func wrapSingleResult[T any](m *T, err error) (*T, *errors.Error) {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Wrap(err, ErrNotFound)
		}
		return nil, errors.Wrap(err, "")
	}

	return m, nil
}
