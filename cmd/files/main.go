// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"

	"internal/app"
)

func main() {
	app := app.NewApp()
	fmt.Println(app, app.Server)
}
