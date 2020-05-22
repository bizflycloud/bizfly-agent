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

package main

import (
	"log"
	"time"

	"git.paas.vn/OpenStack-Infra/bizfly-agent/client"
	"git.paas.vn/OpenStack-Infra/bizfly-agent/cmd"
	"git.paas.vn/OpenStack-Infra/bizfly-agent/collectors"
	"git.paas.vn/OpenStack-Infra/bizfly-agent/config"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/spf13/viper"
)

func init() {
	cmd.InitConfig()
}

func main() {
	var cfg config.Configurations
	viper.Unmarshal(&cfg)
	var httpClient = client.NewHTTPClient()
	httpClient.AuthToken()

	pushGatewayAddress := cfg.PushGW.URL
	waitDuration := cfg.PushGW.WaitDuration

	nc, err := collectors.NewNodeCollector(collectors.DefaultCollectors)
	if err != nil {
		log.Fatalf("failed to create new collector: %s\n", err.Error())
	}

	pusher := push.New(pushGatewayAddress, "bizfly-agent").
		Client(httpClient).
		Grouping("instance_id", cfg.Agent.ID).
		Grouping("hostname", cfg.Agent.Hostname).
		Grouping("instance", cfg.Agent.Hostname).
		Collector(nc)

	if err := pusher.Push(); err != nil {
		log.Fatalf("failed to make initial push to push gateway: %s", err.Error())
	}
	for {
		time.Sleep(time.Second * time.Duration(waitDuration))
		if err := pusher.Push(); err != nil {
			log.Fatalf("failed to push data was collected to push gateway: %s\n", err.Error())
		} else {
			log.Println("pushing data was collected to push gateway")
		}
	}
}
