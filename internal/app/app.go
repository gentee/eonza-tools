// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package app

import (
	"encoding/gob"
	"flag"
	"log"
	"math/rand"
	"os"
	"time"
)

type CmdPar struct {
	Unique uint32
	Ch     chan CmdResult
	Value  interface{}
}

type CmdResult struct {
	Error    error
	Value    interface{}
	Unique   uint32
	Finished bool
}

type CmdFunc func(CmdPar) CmdResult

type CmdHandle struct {
	Func CmdFunc
}

type AppSettings struct {
	Handles map[string]CmdHandle
}

type App struct {
	TaskID   int
	TaskPort int
	Settings AppSettings
	Server   *Server

	Results chan CmdResult
	Latest  time.Time
	Exit    chan int
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewApp(settings AppSettings) *App {
	taskFlag := flag.Int("t", 0, "task id")
	portFlag := flag.Int("p", 0, "port of the task")

	flag.Parse()

	app := App{
		TaskID:   *taskFlag,
		TaskPort: *portFlag,
		Settings: settings,
		Results:  make(chan CmdResult, 7),
		Latest:   time.Now(),
		Exit:     make(chan int),
	}
	log.SetFlags(0)
	if app.TaskID == 0 || app.TaskPort == 0 {
		log.Fatal(`undefined task id or task port`)
	}

	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})

	server, err := NewServer(&app)
	if err != nil {
		log.Fatal(err)
	}
	app.Server = server

	if _, err := SendCmd(app.TaskPort, &CmdData{
		Cmd:    CmdStart,
		TaskID: uint32(app.TaskID),
		Value:  app.Server.Port,
	}); err != nil {
		log.Fatal(err)
	}
	go pingTask(&app)
	return &app
}

func pingTask(app *App) {
	latest := time.Now()
	for {
		time.Sleep(5 * time.Minute)
		if latest.After(app.Latest) {
			if _, err := SendCmd(app.TaskPort, &CmdData{
				Cmd:    CmdPing,
				TaskID: uint32(app.TaskID),
			}); err != nil {
				Shutdown(app)
			}
		}
		latest = time.Now()
	}
}

func Shutdown(app *App) {
	os.Exit(0)
}
