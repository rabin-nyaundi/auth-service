package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/felixge/httpsnoop"
	"rabitech.auth.app/internal/data"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				fmt.Print("ERROR AT PANIC", err)
				app.logger.PrintFatal(errors.New("error: "), nil)
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

func (app *application) requireAdminUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.ContextGetUser(r)

		if user.Role != 1 {
			fmt.Println("user is not admin")
			app.writeJSON(w, http.StatusBadRequest, envelope{"error": "you can't access this resource, !admin"})
			return
		}

		next.ServeHTTP(w, r)
	})
	return app.requireActivatedUser(fn)
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.ContextGetUser(r)
		fmt.Println(user, "user", "user")
		if !user.Active {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return app.requireAuthenticatedUser(fn)
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {

			for i := range app.config.cors.trustedURLOrigins {
				originURL, err := url.Parse(origin)

				if err != nil {
					return
				}

				if originURL == app.config.cors.trustedURLOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {

						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) metrics(next http.Handler) http.Handler {

	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")
	totalResponseSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// start := time.Now()

		totalRequestsReceived.Add(1)

		metrics := httpsnoop.CaptureMetrics(next, w, r)
		// next.ServeHTTP(w, r)

		totalResponsesSent.Add(1)

		// duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())

		totalResponseSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
