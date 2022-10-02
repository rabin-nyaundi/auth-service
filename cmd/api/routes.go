package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/", app.requireActivatedUser(app.status)) // root url to show app status

	router.HandlerFunc(http.MethodPost, "/api/v1/user", app.fetchUserHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/me", app.requireActivatedUser(app.fetchUserHandler))

	router.HandlerFunc(http.MethodGet, "/v1/users", app.requireAdminUser(app.listUsersHandler))         //list all users
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)                           // create account
	router.HandlerFunc(http.MethodPost, "/v1/users/activate", app.activateUserHandler)                  // activate account
	router.HandlerFunc(http.MethodPost, "/v1/token/authenticate", app.createAuthenticationTokenHandler) // login

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler()) // metrics

	return app.metrics(app.recoverPanic(app.enableCORS(app.authenticate(router))))
}
