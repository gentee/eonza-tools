// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"time"

	chi "github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Port int
}

func defaultHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func shutdownHandle(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value("app").(*App)

	if app.TaskID == StrToInt(chi.URLParam(r, "TaskID")) {
		app.Exit <- 0
	}
}

func NewServer(app *App) (*Server, error) {
	var (
		server Server
		err    error
	)
	port := app.TaskPort + 1000

	ch := make(chan error, 1)
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "app", app)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	r.Get("/", defaultHandle)
	r.Get("/shutdown/{TaskID}", shutdownHandle)

	startServer := func() {
		ch <- http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	}

start:
	for {
		if port, err = GetFreePort(port); err != nil {
			return nil, err
		}
		go startServer()
		select {
		case err = <-ch:
			// too fast error. Probably "bind: address already in use"
			// try to change port
		case <-time.After(50 * time.Millisecond):
			break start
		}
	}
	server.Port = port
	return &server, nil
}
