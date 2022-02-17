// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package app

import (
	"flag"
	"fmt"
	"log"
)

type App struct {
	TaskID   int
	TaskPort int
	Port     int
	Server   *Server

	Exit chan int
}

func NewApp() *App {
	taskFlag := flag.Int("t", 0, "task id")
	portFlag := flag.Int("p", 0, "port of the task")

	flag.Parse()

	app := App{
		TaskID:   *taskFlag,
		TaskPort: *portFlag,
		Exit:     make(chan int),
	}

	if app.TaskID == 0 || app.TaskPort == 0 {
		log.Fatal(`undefined task id or task port`)
	}

	server, err := NewServer(&app)
	if err != nil {
		log.Fatal(err)
	}
	app.Server = server

	if body, err := LocalGet(app.TaskPort, fmt.Sprintf("pkg/%d", app.Port)); err != nil {
		log.Fatal(err)
	} else if StrToInt(string(body)) != app.TaskID {
		log.Fatal(`wrong task id`)
	}

	return &app
}
