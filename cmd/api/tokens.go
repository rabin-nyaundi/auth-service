package main

import (
	"fmt"
	"net/http"
	"time"

	"rabitech.auth.app/internal/data"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		return
	}

	user, err := app.models.User.GetUserByEmail(input.Email)

	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		return
	}

	if !user.Active {
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": "account not activated"})
		return
	}

	fmt.Println(input.Password, user.Password)

	passwordMatch, err := user.Password.MatchPassword(input.Password)

	if err != nil {
		fmt.Println("Error here")
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		return
	}

	if !passwordMatch {
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": "passwords doesn't match"})
		return
	}

	token, err := app.models.Tokens.New(user.ID, 1*24*time.Hour, data.ScopeAuthentication)

	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token})
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})
		return
	}

}
