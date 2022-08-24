package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/", app.status)
	router.HandlerFunc(http.MethodGet, "/v1/users", app.listUsersHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users/activated", app.activateUserHandler)

	return app.recoverPanic(router)
}
