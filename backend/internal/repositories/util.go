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

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func fromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func toNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func fromNullInt64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	v := ni.Int64
	return &v
}

func toNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func fromNullFloat64(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	v := nf.Float64
	return &v
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
