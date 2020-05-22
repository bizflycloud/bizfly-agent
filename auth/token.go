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

	"git.paas.vn/OpenStack-Infra/bizfly-agent/config"
	"github.com/spf13/viper"
)

var (
	cfg                     config.Configurations
	defaultMetadataEndpoint = cfg.AuthServer.DefaultMetadataEndpoint
)

// Token is exported
type Token struct {
}

// SaveToken is exported
func (t *Token) SaveToken(token string) {
	viper.Unmarshal(&cfg)

	// log.Println(cfg.ConfigDir)
	file, err := os.Create(cfg.ConfigDir + "/auth_token")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Saving auth token")
	_, err = file.WriteString(token)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
}

// ReadToken is exported
func (t *Token) ReadToken() (string, error) {
	viper.Unmarshal(&cfg)

	log.Println("Reading auth token")
	data, err := ioutil.ReadFile(cfg.ConfigDir + "/auth_token")
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return string(data), nil
}
