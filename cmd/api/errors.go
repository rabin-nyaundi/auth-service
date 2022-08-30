package main

import (
	"net/http"
)

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"

	app.writeJSON(w, http.StatusUnauthorized, envelope{"error": message})
}

func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.writeJSON(w, http.StatusForbidden, envelope{"error": message})
}

func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your account must be activated to access the resource"
	app.writeJSON(w, http.StatusForbidden, envelope{"error": message})
}
