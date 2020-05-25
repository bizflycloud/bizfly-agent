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

	"github.com/bizflycloud/bizfly-agent/config"
)

const authTokenFilename = "auth_token"

// Token ...
type Token struct{}

// SaveToken is save auth token to file auth_token in config directory
// Client will use this token instead by get new token every time.
func (t *Token) SaveToken(token string) error {
	filename := filepath.Join(config.Config.ConfigDir, authTokenFilename)

	file, err := os.Create(filename)
	if err != nil {
		prol.Fatal(err)
	}

	prol.Debugln("Saving auth token to: ", filename)
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
	data, err := ioutil.ReadFile(filepath.Join(config.Config.ConfigDir, authTokenFilename))
	if err != nil {
		prol.Fatal(err)
		return "", err
	}
	return string(data), nil
}
