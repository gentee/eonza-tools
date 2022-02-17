// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package server

import (
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

func NewServer(port int) (*Server, error) {
	var (
		server Server
		err    error
	)

	ch := make(chan error, 1)
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/", defaultHandle)

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
