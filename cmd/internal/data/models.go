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

//  NewModel function initailize a new model
func NewModel(db *sql.DB) Models {
	return Models{
		User:   UserModel{DB: db},
		Tokens: TokenModel{DB: db},
	}
}
