// Copyright 2022 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"internal/app"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func IsDir(fname string) (bool, error) {
	var (
		fi  os.FileInfo
		err error
	)
	if fi, err = os.Stat(fname); err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

/*func IsEmptyDir(path string) (bool, error) {
	var (
		f   *os.File
		err error
	)

	if f, err = os.Open(path); err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.Readdir(1)

	if err == io.EOF {
		return true, nil
	}
	return false, err
}*/

// MatchºStrStr reports whether the string s contains any match of the regular expression
func MatchReg(s string, rePattern string) (bool, error) {
	re, err := regexp.Compile(rePattern)
	if err != nil {
		return false, err
	}
	if re.MatchString(s) {
		return true, nil
	}
	return false, nil
}

// MatchPath reports whether name matches the specified file name pattern.
func MatchPath(pattern, fname string) (bool, error) {
	if len(pattern) == 0 {
		return true, nil
	}
	var (
		ok      bool
		isRegex bool
		err     error
	)
	if len(pattern) > 2 && pattern[0] == '/' && pattern[len(pattern)-1] == '/' {
		isRegex = true
		pattern = pattern[1 : len(pattern)-1]
	}
	if isRegex {
		return MatchReg(fname, pattern)
	} else {
		ok, err = filepath.Match(pattern, fname)
		if ok {
			return true, err
		}
	}
	return false, err
}

func searchEmptyDirs(root string, ch chan string, patterns, ignores []string) error {
	dir, err := os.Open(root)
	if err != nil {
		return err
	}
	defer dir.Close()
	dirs, err := dir.ReadDir(0)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		name := filepath.Base(root)
		if len(ignores) > 0 {
			if matched, err := MatchPatterns(ignores, name); err != nil {
				return err
			} else if matched {
				return nil
			}
		}
		if len(patterns) > 0 {
			if matched, err := MatchPatterns(patterns, name); err != nil {
				return err
			} else if !matched {
				return nil
			}
		}
		ch <- root
		return nil
	}
	for _, d := range dirs {
		if d.IsDir() {
			if err = searchEmptyDirs(filepath.Join(root, d.Name()), ch, patterns, ignores); err != nil {
				return err
			}
		}
	}
	return nil
}

func MatchPatterns(patterns []string, name string) (matched bool, err error) {
	if len(patterns) > 0 {
		for _, v := range patterns {
			if matched, err = MatchPath(v, name); err != nil {
				return
			}
			if matched {
				break
			}
		}
	}
	return
}

func SplitPar(par map[string]interface{}, key, sep string) (pars []string) {
	pattern := fmt.Sprint(par[key])
	if len(pattern) > 0 {
		pars = strings.Split(pattern, sep)
		for i, v := range pars {
			pars[i] = strings.TrimSpace(v)
		}
	}
	return
}

func EmptyDirs(cmd app.CmdPar) app.CmdResult {
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
	ch := make(chan string, 7)
	go func() {
		err = searchEmptyDirs(root, ch, patterns, ignores)
		close(ch)
	}()
	Limit := 30
	list := make([]string, Limit)
	limit := 0
	for s := range ch {
		list[limit] = s
		limit++
		if limit == Limit {
			result.Value = list
			cmd.Ch <- result
			limit = 0
		}
	}
	if err != nil {
		return errResult(err)
	}
	result.Value = list[:limit]
	return result
}
