package db

import (
	"database/sql"
	"time"
)

var (
	NullTime = func(input time.Time) sql.NullTime {
		return sql.NullTime{Time: input, Valid: true}
	}
	NullBigint = func(input uint64) sql.Null[uint64] {
		if input == 0 {
			return sql.Null[uint64]{Valid: false}
		}
		return sql.Null[uint64]{V: input, Valid: true}
	}
)
