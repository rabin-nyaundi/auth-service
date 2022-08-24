package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"fmt"
	"time"
)

/*
Defne constant for the scope
*/
const (
	ScopeActivation = "activation"
)

/*
Token struct to hold the data for individual token.
It includes plaintext and hashed plaintext, id of the associated user, time for the token to expire and scope of the token.
*/
type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int
	Expiry    time.Time
	Scope     string
}

/*
Token model
*/
type TokenModel struct {
	DB *sql.DB
}

/*
New generates new token and inserts into database,
given associated user's id, duration for expiry and scope of the token.
*/
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := GenerateToken(int64(userID), ttl, scope)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	err = m.Insert(token)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return token, nil
}

/*
Make a query to insert the token into database table
*/
func (m TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2,$3,$4)
	RETURNING user_id
	`

	args := []interface{}{
		&token.Hash, token.UserID, &token.Expiry, &token.Scope,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&token.UserID)

	if err != nil {
		fmt.Println(err)
	}

	return err
}

/*
DelteAllForUser() deletes all tokens for user and the scope from tokens table
*/
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 and user_id = $2
	RETURNING user_id`

	args := []interface{}{
		scope, userID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, args...)

	return err
}

/*
 GenerateToken generates a nw token.
 Takes userid argument, ttl duration for the token to expire and scope of the token.
*/
func GenerateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {

	token := &Token{
		UserID: int(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)

	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

/*
GetUserForToken retrieves a user associated with a token.
*/

/*
GetAllTokenForUser retrieves all token associated with a user.
*/