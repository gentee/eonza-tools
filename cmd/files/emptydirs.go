// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"internal/app"
)

func EmptyDirs(cmd app.CmdPar) app.CmdResult {
	result := app.CmdResult{}
	if par, ok := cmd.Value.(map[string]interface{}); !ok {
		result.Error = fmt.Errorf(`wrong cmd parameter`)
	} else {
		result.Value = fmt.Sprintf(`Empty dirs %v`, par)
	}
	return result
}
