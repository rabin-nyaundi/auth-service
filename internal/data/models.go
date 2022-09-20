package data

import (
	"database/sql"
	"errors"
)

//  ErrorRecordNotFound record not found error
var (
	ErrorRecordNotFound = errors.New("record not found")
)

// Models struct
type Models struct {
	User   UserModel
	Tokens TokenModel
}

//  NewModel return models.
func NewModel(db *sql.DB) Models {
	return Models{
		User:   UserModel{DB: db},
		Tokens: TokenModel{DB: db},
	}
}
