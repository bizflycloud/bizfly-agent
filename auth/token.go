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
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/bizflycloud/bizfly-agent/config"
)

var (
	cfg                     config.Configurations
	defaultMetadataEndpoint = cfg.AuthServer.DefaultMetadataEndpoint
)

// Token ...
type Token struct {
}

// SaveToken is save auth token to file auth_token in config directory
// Client will use this token instead by get new token every time.
func (t *Token) SaveToken(token string) error {
	viper.Unmarshal(&cfg)

	file, err := os.Create(cfg.ConfigDir + "/auth_token")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Saving auth token")
	_, err = file.WriteString(token)
	if err != nil {
		log.Fatal(err)
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}

// ReadToken is read auth token was saved before
func (t *Token) ReadToken() (string, error) {
	viper.Unmarshal(&cfg)

	log.Println("Reading auth token")
	data, err := ioutil.ReadFile(filepath.Join(cfg.ConfigDir, "/auth_token"))
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return string(data), nil
}
