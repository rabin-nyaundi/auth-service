package main

import (
	"context"
	"net/http"

	"rabitech.auth.app/cmd/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) ContextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)

	return r.WithContext(ctx)
}

func (app *application) ContextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user key value in context")
	}
	return user
}
