package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DBTransaction struct {
	tx *gorm.DB
}

func NewTransaction(db *gorm.DB) *DBTransaction {
	trx := &DBTransaction{tx: db}
	trx.begin(db)
	return trx
}

func (t *DBTransaction) begin(db *gorm.DB) {
	t.tx = db.Begin()
}

func GetRepo[T schema.Tabler](t *DBTransaction, dm T) *GenericRepository[T] {
	return NewGenericeRepoPointr(t.tx, dm)
}

func (t *DBTransaction) End(err error) error {
	if err != nil {
		return t.tx.Rollback().Error
	}
	return t.tx.Commit().Error
}
