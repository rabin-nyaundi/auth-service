package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ErrorDuplicateEmail returned when duplicate email is provided
var (
	ErrorDuplicateEmail = errors.New("duplicate email")
	// ErrorRecordNotFound = errors.New("record not found")
)

// AnonymusUser is a non user not existing in our system
var AnonymusUser = &User{}

// User struct with user properties
type User struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"firstname"`
	LastName  string    `json:"lastname"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Active    bool      `json:"active"`
	Role      int       `json:"role"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

// UserModel struct fot user model
type UserModel struct {
	DB *sql.DB
}

func (u *User) IsAnonymus() bool {
	return u == AnonymusUser
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)

	if err != nil {
		log.Fatal(err)
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) MatchPassword(plaintextPassword string) (bool, error) {

	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}

	}

	return true, nil
}

func (m UserModel) InsertUser(user *User) error {
	query := `
	INSERT INTO auth_user (firstname, lastname, email, username, password_hash, active, role, version)
	VALUES ($1, $2, $3, $4, $5, $6, 0, 1)
	RETURNING id, CreatedAt, version
	`

	args := []interface{}{
		user.FirstName, user.LastName, user.Email, user.Username, user.Password.hash, user.Active,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fmt.Println("This code is reached")

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "auth_user_email_key"`:
			return ErrorDuplicateEmail
		default:
			fmt.Println("I have an eror here. Kindly fix it")
			return err
		}
	}
	return nil
}

func (m UserModel) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, firstname, lastname, username, email, password_hash, active
		FROM auth_user
		WHERE email = $1`
	var user User
	args := []interface{}{
		email,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Active,
	)

	if err != nil {
		switch {
		case err.Error() == `sql: no rows in result set`:
			return nil, ErrorRecordNotFound
		default:
			fmt.Println(err.Error())
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) UpdateUser(user *User) error {
	query := `
		UPDATE auth_user
		SET active = $1, version = $2, UpdatedAt = $3
		WHERE id = $4
		RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{
		user.Active, user.Version, user.UpdatedAt, user.ID,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (m UserModel) GetUsers() ([]*User, error) {
	query := `SELECT * FROM auth_user`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fmt.Println("Get all users hit")
	rows, err := m.DB.QueryContext(ctx, query)
	fmt.Println("Get all users hit again")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Password.hash,
			&user.Username,
			&user.Active,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Version,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("Error here")
		return nil, err
	}

	return users, nil

}

/*
GetUserForToken retrieves a user associated with a token.
*/
func (m UserModel) GetUserForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `SELECT auth_user.id, auth_user.firstname, auth_user.lastname, auth_user.username,auth_user.CreatedAt, auth_user.email, auth_user.password_hash, auth_user.active, auth_user.role
		FROM auth_user
		INNER JOIN tokens
		ON auth_user.id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.scope = $2
		AND tokens.expiry > $3`

	args := []interface{}{
		tokenHash[:], tokenScope, time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.CreatedAt,
		&user.Email,
		&user.Password.hash,
		&user.Active,
		&user.Role,
	)

	if err != nil {
		switch {
		case err.Error() == `sql: no rows in result set`:
			return nil, ErrorRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}

/*
GetAllTokenForUser retrieves all token associated with a user.
*/
