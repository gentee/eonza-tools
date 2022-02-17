// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package app

import (
	"flag"
	"log"

	"internal/server"
)

type App struct {
	TaskID   uint32
	TaskPort int
	Port     int
	Server   *server.Server
}

func NewApp() *App {
	var app App

	taskFlag := flag.Int("t", 0, "task id")
	portFlag := flag.Int("p", 0, "port of the task")

	flag.Parse()

	app.TaskID = uint32(*taskFlag)
	app.TaskPort = *portFlag

	if app.TaskID == 0 || app.TaskPort == 0 {
		log.Fatal(`undefined task id or task port`)
	}

	server, err := server.NewServer(app.TaskPort + 1000)
	if err != nil {
		log.Fatal(err)
	}
	app.Server = server
	return &app
}
