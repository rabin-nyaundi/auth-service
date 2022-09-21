package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (app *application) writeJSON(w http.ResponseWriter, status int, data any) error {
	jsonObject, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		app.logger.PrintError(err, map[string]string{
			"error": "failed to write json",
		})
		fmt.Println(err)
	}

	jsonObject = append(jsonObject, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonObject)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalerror *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field == "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request too large":
			return fmt.Errorf("body must not be larger that %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalerror):
			panic(err)

		default:
			fmt.Println("Heeey")
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return fmt.Errorf("body must contain a single JSON object")
	}

	return nil
}

func (app *application) background(fn func()) {

	app.wg.Add(1)
	go func() {

		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		fn()
	}()
}

func (app *application) JSONError(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	return app.writeJSON(w, statusCode,
		JSONResponse{
			Error:   true,
			Message: err.Error(),
		})

}
