package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorDuplicateEmail = errors.New("duplicate email")
)

type User struct {
	ID         int64     `json:"id"`
	FirstName  string    `json:"firstname"`
	LastName   string    `json:"lastname"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Active     bool      `json:"active"`
	Role       int       `json:"role"`
	Created_At time.Time `json:"created_at"`
	Updated_At time.Time `json:"updated_at"`
	Version    int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

type UserModel struct {
	DB *sql.DB
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
	VALUES ($1, $2, $3, $4, $5, $6, 1, 1)
	RETURNING id, created, version
	`

	args := []interface{}{
		user.FirstName, user.LastName, user.Email, user.Username, user.Password.hash, user.Active,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fmt.Println("This code is reached")

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID)

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
			&user.Created_At,
			&user.Updated_At,
			&user.Version,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil

}
