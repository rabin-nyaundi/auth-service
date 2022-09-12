package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"rabitech.auth.app/cmd/internal/data"
)

func (app *application) status(w http.ResponseWriter, r *http.Request) {
	status := envelope{
		"status":      "available",
		"Version":     version,
		"Environment": app.config.env,
	}

	err := app.writeJSON(w, http.StatusOK, status)

	if err != nil {
		fmt.Println(err)
	}
}

func (app *application) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := app.models.User.GetUsers()

	if err != nil {
		fmt.Println(err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"users": users})

	if err != nil {
		fmt.Println(err)
	}
}

func (app *application) fetchUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		TokenPlaintext string `json:"token"`
	}
	err := app.readJSON(w, r, &input)

	if err != nil {
		fmt.Println(err)
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		return
	}

	user, err := app.models.User.GetUserForToken(data.ScopeAuthentication, input.TokenPlaintext)

	if err != nil {
		switch {
		case err.Error() == `sql: no rows in result set`:
			app.writeJSON(w, http.StatusBadRequest, envelope{"error": "no user found with such token"})
			return
		default:
			fmt.Println(err)
			app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		}
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"data": user})
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		fmt.Println(err)
		return
	}

	user := &data.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Username:  input.Username,
		Email:     input.Email,
		Active:    false,
	}

	err = user.Password.Set(input.Password)

	if err != nil {
		fmt.Println(err)
		return
	}

	err = app.models.User.InsertUser(user)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrorDuplicateEmail):
			fmt.Println("duplicate email found")
			app.writeJSON(w, http.StatusBadRequest, envelope{"error": "user with email already exist."})

		default:
			fmt.Println(err)
			fmt.Println("Error inserting user to database")
			app.writeJSON(w, http.StatusBadRequest, envelope{"error": "Error inserting user to database"})
		}
		return
	}
	duration := 1 * 24 * time.Hour

	token, err := app.models.Tokens.New(user.ID, duration, data.ScopeActivation)
	if err != nil {
		fmt.Println(err)
		return
	}

	emailData := map[string]interface{}{
		"UserID":          user.ID,
		"UserName":        user.Username,
		"activationToken": token.Plaintext,
		"expiryDuration":  duration,
	}
	app.background(func() {
		err = app.mailer.Send(user.Email, "user_registration.html", emailData)

		if err != nil {
			switch {
			case err.Error() == "template: pattern matches no files: `templates/user_registration.tmpl`":
				app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
				return
			default:
				fmt.Println(err)
				app.writeJSON(w, http.StatusBadRequest, envelope{"error": "an error occured when sending activation email"})
			}
			return
		}
	})

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": "user created successfuly"})

	if err != nil {
		fmt.Println(err)
		return
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		fmt.Println(err)
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		return
	}

	user, err := app.models.User.GetUserForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case err.Error() == "sql: no rows in result set":
			app.writeJSON(w, http.StatusBadRequest, envelope{"error": "no user found with that token"})
			return

		default:
			app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
			fmt.Println(err, "error at get user for a token")
		}
		return
	}

	user.Active = true
	user.UpdatedAt = time.Now()
	user.Version = user.Version + 1

	err = app.models.User.UpdateUser(user)

	if err != nil {
		fmt.Println(err, "Error at Updat user")
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		fmt.Println(err, "error at delete user tokens")
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		return
	}

	app.writeJSON(w, http.StatusAccepted, envelope{"data": user})
}
