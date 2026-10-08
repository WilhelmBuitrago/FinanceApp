package repositories

import (
	"database/sql"
	"errors"
	"strings"
)

func toInterface(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func toInterfaceInt64(i *int64) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

func toInterfaceFloat64(f *float64) interface{} {
	if f == nil {
		return nil
	}
	return *f
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "unique constraint failed") || strings.Contains(msg, "constraint failed") {
		if strings.Contains(msg, "unique") {
			return ErrConflict
		}
		return ErrConstraint
	}
	if strings.Contains(msg, "no such table") || strings.Contains(msg, "database is locked") {
		return ErrUnavailable
	}
	return ErrInternal
}
