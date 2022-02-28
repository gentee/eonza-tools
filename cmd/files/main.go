// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"os"

	"internal/app"
)

var cmds = map[string]app.CmdHandle{
	`emptydirs`: {EmptyDirs},
}

func main() {
	app := app.NewApp(app.AppSettings{
		cmds,
	})

	os.Exit(<-app.Exit)
}
