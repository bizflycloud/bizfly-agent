// This file is part of bizfly-agent
//
// Copyright (C) 2020  BizFly Cloud
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>

package auth

import (
	"io/ioutil"
	"os"
	"path/filepath"

	prol "github.com/prometheus/common/log"
)

const authTokenFilename = "auth_token"

// Token ...
type Token struct {
	authTokenFile string
}

// NewToken returns new Token instance.
func NewToken() (*Token, error) {
	t := &Token{}
	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	cfgDir := filepath.Join(userCfgDir, "bizfly-agent")
	if err := os.MkdirAll(cfgDir, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}

	t.authTokenFile = filepath.Join(cfgDir, authTokenFilename)
	return t, nil
}

// SaveToken is save auth token to file auth_token in config directory
// Client will use this token instead by get new token every time.
func (t *Token) SaveToken(token string) error {
	file, err := os.Create(t.authTokenFile)
	if err != nil {
		prol.Fatal(err)
	}
	err = file.Chmod(0600)
	if err != nil {
		prol.Errorln("Can't change mod of file auth_token")
	}

	prol.Debugln("Saving auth token to: ", t.authTokenFile)
	_, err = file.WriteString(token)
	if err != nil {
		prol.Fatal(err)
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}

// ReadToken is read auth token was saved before
func (t *Token) ReadToken() (string, error) {
	prol.Debugln("Reading auth token")
	data, err := ioutil.ReadFile(t.authTokenFile)
	if err != nil {
		prol.Fatal(err)
		return "", err
	}
	return string(data), nil
}
