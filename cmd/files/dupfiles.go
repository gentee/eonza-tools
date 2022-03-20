// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"internal/app"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func MD5Hash(fname string) (string, error) {
	file, err := os.Open(fname)
	if err != nil {
		return ``, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ``, err
	}
	return hex.EncodeToString(hash.Sum(nil)[:]), nil
}

func searchFiles(root string, list map[int64][]string, patterns, ignores []string) error {
	dir, err := os.Open(root)
	if err != nil {
		return err
	}
	defer dir.Close()
	dirs, err := dir.ReadDir(0)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			if err = searchFiles(filepath.Join(root, name), list, patterns, ignores); err != nil {
				return err
			}
		} else {
			if len(ignores) > 0 {
				if matched, err := MatchPatterns(ignores, name); err != nil {
					return err
				} else if matched {
					continue
				}
			}
			if len(patterns) > 0 {
				if matched, err := MatchPatterns(patterns, name); err != nil {
					return err
				} else if !matched {
					continue
				}
			}
			if fi, err := d.Info(); err != nil {
				return err
			} else {
				size := fi.Size()
				if size > 0 {
					list[size] = append(list[size], filepath.Join(root, name))
				}
			}
		}
	}
	return nil
}

func DupFiles(cmd app.CmdPar) app.CmdResult {
	var (
		par map[string]interface{}
		ok  bool
	)
	result := app.CmdResult{Unique: cmd.Unique}

	errResult := func(err error) app.CmdResult {
		result.Error = err
		return result
	}

	if par, ok = cmd.Value.(map[string]interface{}); !ok {
		return errResult(fmt.Errorf(`wrong cmd parameter`))
	}
	root, err := filepath.Abs(fmt.Sprint(par["path"]))

	patterns := SplitPar(par, "pattern", ",")
	ignores := SplitPar(par, "ignore", ",")

	if isdir, err := IsDir(root); err != nil {
		return errResult(err)
	} else if !isdir {
		return errResult(fmt.Errorf(`'%s' is not a directory`, root))
	}
	fileList := make(map[int64][]string)
	err = searchFiles(root, fileList, patterns, ignores)
	if err != nil {
		return errResult(err)
	}
	sizes := make([]int64, 0, 1024)
	for size, v := range fileList {
		if len(v) > 1 {
			sizes = append(sizes, size)
		}
	}
	sort.Slice(sizes, func(i, j int) bool { return sizes[i] < sizes[j] })
	for _, size := range sizes {
		hash := make(map[string][]int)
		for i, f := range fileList[size] {
			md5, err := MD5Hash(f)
			if err != nil {
				return errResult(err)
			}
			hash[md5] = append(hash[md5], i)
		}
		for key, list := range hash {
			if len(list) > 1 {
				files := make([]string, len(list))
				for i, v := range list {
					files[i] = fileList[size][v]
				}
				result.Value = map[string]interface{}{
					"size": size,
					"md5":  key,
					"list": files,
				}
				cmd.Ch <- result
			}
		}
	}
	result.Value = 0
	return result
}
