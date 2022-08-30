package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"rabitech.auth.app/cmd/internal/data"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				fmt.Print(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.ContextSetUser(r, data.AnonymusUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		user, err := app.models.User.GetUserForToken(data.ScopeAuthentication, token)

		if err != nil {
			switch {
			case errors.Is(err, data.ErrorRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()})

			}
			return
		}
		r = app.ContextSetUser(r, user)
		next.ServeHTTP(w, r)

		fmt.Println(app.ContextGetUser(r).Email)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.ContextGetUser(r)

		if user.IsAnonymus() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.ContextGetUser(r)
		fmt.Println(user,"user", "usererre")
		if !user.Active {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return app.requireAuthenticatedUser(fn)
}
