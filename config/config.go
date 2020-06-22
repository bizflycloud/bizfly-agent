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

package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// Config is the global configuration.
var Config Configurations
var o sync.Once

func init() {
	o.Do(func() {
		viper.SetConfigName("bizfly-agent")
		viper.SetConfigType("yaml")

		userCfgDir, err := os.UserConfigDir()
		if err != nil {
			panic("Can't get user config directory")
		}
		cfgDir := filepath.Join(userCfgDir, "bizfly-agent")
		viper.AddConfigPath(cfgDir)

		viper.AddConfigPath("/etc/bizfly-agent")
		viper.AddConfigPath(".")

		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}

		hostname, err := os.Hostname()
		if err != nil {
			panic("Can't get hostname.")
		}

		// Config.Agent.Hostname = hostname
		// Config.Agent.Name = hostname

		// Set config
		viper.Set("agent.name", hostname)
		viper.Set("agent.hostname", hostname)

		if err := viper.Unmarshal(&Config); err != nil {
			panic(err)
		}
	})
}

// Configurations contains all configuration.
type Configurations struct {
	Agent      AgentsConfigurations
	AuthServer ServersConfigurations
	PushGW     PushGateWay
}

// AgentsConfigurations is agent configuration.
type AgentsConfigurations struct {
	ID       string
	Name     string
	Hostname string
}

// ServersConfigurations contains server configuration.
type ServersConfigurations struct {
	DefaultEndpoint string
	Secret          string
	SecretID        string
	Project         string
}

// PushGateWay contains push gateway configuration.
type PushGateWay struct {
	URL          string
	WaitDuration int
}
