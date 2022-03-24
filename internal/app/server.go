// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package app

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	chi "github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Port int
}

func cmdHandle(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value("app").(*App)

	var (
		response CmdData
		cmd      *CmdData
		err      error
	)

	if cmd, err = ProcessCmd(uint32(app.TaskID), r.Body); err != nil {
		response.Error = err.Error()
	} else {
		switch cmd.Cmd {
		case CmdShutdown:
			Shutdown(app)
		default:
			if handle, ok := app.Settings.Handles[cmd.Cmd]; ok {
				go func() {
					result := handle.Func(CmdPar{Value: cmd.Value, Ch: app.Results, Unique: cmd.Unique})
					result.Finished = true
					result.Unique = cmd.Unique
					app.Results <- result
				}()
			} else {
				response.Error = fmt.Sprintf(`unknown cmd "%s"`, cmd.Cmd)
			}
		}
	}
	response.TaskID = uint32(app.TaskID)

	w.Write(ResponseCmd(&response))
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
			app.Latest = time.Now()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	r.Post("/cmd", cmdHandle)

	startServer := func() {
		ch <- http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	}

	go func() {
		var data bytes.Buffer

		for {
			result := <-app.Results
			data.Reset()
			enc := gob.NewEncoder(&data)
			answer := CmdData{
				TaskID:   uint32(app.TaskID),
				Unique:   result.Unique,
				Value:    result.Value,
				Finished: result.Finished,
			}
			if result.Error != nil {
				answer.Error = result.Error.Error()
			}
			if err = enc.Encode(answer); err == nil {
				resp, err := http.Post(fmt.Sprintf("http://localhost:%d/cmdresult", app.TaskPort),
					"application/octet-stream", &data)
				if err == nil {
					resp.Body.Close()
				}
			}
		}
	}()

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
