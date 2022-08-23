package data

import (
	"database/sql"
	"errors"
)

var (
	ErrorRecordNotFound = errors.New("record not found")
)

type Models struct {
	User UserModel
}

func NewModel(db *sql.DB) Models {
	return Models{
		User: UserModel{DB: db},
	}
}
