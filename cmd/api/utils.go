package main

import (
	"context"
	"net/http"

	"github.com/windevkay/flho/internal/data"
)

// custom type for adding keys to request context
type contextKey string

const (
	userContextKey = contextKey("user")
)

// func (app *application) background(fn func()) {
// 	app.wg.Add(1)

// 	go func() {
// 		defer app.wg.Done()

// 		defer func() {
// 			if err := recover(); err != nil {
// 				app.logger.Error(fmt.Sprintf("%v", err))
// 			}
// 		}()

// 		fn()
// 	}()
// }

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
