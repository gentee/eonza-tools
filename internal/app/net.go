// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package app

import (
	"fmt"
	"io"
	"net"
	"net/http"
)

const PortsPool = 1000

func GetFreePort(port int) (int, error) {
	var i int

	for ; i < PortsPool; i++ {
		port++
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			ln.Close()
			break
		}
	}
	if i == PortsPool {
		return i, fmt.Errorf(`There is not available port in the pool`)
	}
	return port, nil
}

func LocalGet(port int, url string) (body []byte, err error) {
	var (
		res *http.Response
	)
	res, err = http.Get(fmt.Sprintf(`http://localhost:%d/%s`, port, url))
	if err == nil {
		body, err = io.ReadAll(res.Body)
		res.Body.Close()
	}
	return
}
