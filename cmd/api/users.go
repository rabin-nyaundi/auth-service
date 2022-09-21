package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"rabitech.auth.app/internal/data"
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

	err = app.writeJSON(w, http.StatusOK,
		JSONResponse{
			Success: true,
			Message: "users fetch success",
			Data:    users,
		})

	if err != nil {
		fmt.Println("handler err")
		fmt.Println(err)
	}
}

func (app *application) fetchUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		TokenPlaintext string `json:"token"`
	}
	err := app.readJSON(w, r, &input)

	if err != nil {
		app.JSONError(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.models.User.GetUserForToken(data.ScopeAuthentication, input.TokenPlaintext)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrorRecordNotFound):
			app.JSONError(w, errors.New("no user found with such token"), http.StatusBadRequest)
			return
		default:
			app.JSONError(w, err, http.StatusBadRequest)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "users fetch success",
		Data:    user,
	})
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
		app.logError(r, err)
		app.JSONError(w, err, http.StatusBadRequest)
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
		app.logError(r, err)
		app.JSONError(w, err, http.StatusBadRequest)
		return
	}

	err = app.models.User.InsertUser(user)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrorDuplicateEmail):
			app.logError(r, err)
			app.JSONError(w, errors.New("user with email already exist"), http.StatusBadRequest)
		default:
			app.logError(r, err)
			app.JSONError(w, errors.New("error inserting user to database"), http.StatusBadRequest)
		}
		return
	}
	duration := 1 * 24 * time.Hour

	token, err := app.models.Tokens.New(user.ID, duration, data.ScopeActivation)
	if err != nil {
		app.logError(r, err)
		app.JSONError(w, err, http.StatusBadRequest)
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
				app.JSONError(w, errors.New("no template found"), http.StatusBadRequest)
				app.logError(r, err)
				return
			default:
				app.JSONError(w, errors.New("an error occured when sending activation email"), http.StatusBadRequest)
				app.logError(r, err)
			}
			return
		}
	})

	err = app.writeJSON(w, http.StatusCreated,
		JSONResponse{
			Success: true,
			Message: "user successfully created",
		})
	if err != nil {
		app.JSONError(w, err, http.StatusBadRequest)
		app.logError(r, err)
		return
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.JSONError(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.models.User.GetUserForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrorRecordNotFound):
			app.JSONError(w, errors.New("no user found match with to token"), http.StatusBadRequest)
			return

		default:
			app.JSONError(w, err, http.StatusBadRequest)
		}
		return
	}

	user.Active = true
	user.UpdatedAt = time.Now()
	user.Version = user.Version + 1

	err = app.models.User.UpdateUser(user)

	if err != nil {
		app.JSONError(w, err, http.StatusBadRequest)
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.JSONError(w, err, http.StatusBadRequest)
		return
	}

	app.writeJSON(w, http.StatusAccepted,
		JSONResponse{
			Success: true,
			Message: "user activation success",
			Data:    user,
		})
}
