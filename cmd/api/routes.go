package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/", app.requireActivatedUser(app.status))
	router.HandlerFunc(http.MethodPost, "/api/v1/user", app.fetchUserHandler)
	// router.HandlerFunc(http.MethodPost, "/api/v1/user", app.requireActivatedUser(app.fetchUserHandler))
	router.HandlerFunc(http.MethodGet, "/v1/users", app.requireAdminUser(app.listUsersHandler))
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/token/authenticate", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.authenticate(router))))
}
