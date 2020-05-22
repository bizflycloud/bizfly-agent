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

package cmd

import (
	"log"

	"git.paas.vn/OpenStack-Infra/bizfly-agent/config"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/spf13/viper"
)

var (
	cfgFile = kingpin.Flag("config-file", "config file (default is /etc/bizfly-agent/bizfly-agent.yaml)").Default("/etc/bizfly-agent/bizfly-agent.yaml").String()
)

// InitConfig reading config from command line
func InitConfig() {
	kingpin.Parse()
	log.Println("Loading config file ...")

	if *cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(*cfgFile)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Fatalln("Unreadable config file:", viper.ConfigFileUsed())
	}

	var Cfg config.Configurations
	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("Unable to decode into struct, %s", err)
	}
}
