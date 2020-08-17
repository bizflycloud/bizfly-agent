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
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"
	prol "github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bizflycloud/bizfly-agent/client"
	"github.com/bizflycloud/bizfly-agent/collectors"
	"github.com/bizflycloud/bizfly-agent/config"
)

func main() {
	// Do not remove theses lines, prometheus needs them to run.
	prol.AddFlags(kingpin.CommandLine)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var httpClient = client.NewHTTPClient()
	if _, err := httpClient.AuthToken(); err != nil {
		prol.Errorf("failed to get client auth token: %s", err)
	}

	pushGatewayAddress := config.Config.PushGW.URL
	waitDuration := config.Config.PushGW.WaitDuration

	nc, err := collectors.NewNodeCollector(collectors.DefaultCollectors)
	if err != nil {
		prol.Fatalf("failed to create new collector: %s\n", err.Error())
	}

	pusher := push.New(pushGatewayAddress, "bizfly-agent").
		Client(httpClient).
		Grouping("hostname", config.Config.Agent.Hostname).
		Grouping("instance", config.Config.Agent.Name).
		Grouping("instance_id", config.Config.Agent.ID).
		Grouping("project_id", config.Config.AuthServer.Project).
		Grouping("runtime", runtime.GOOS).
		Collector(nc)

	if err := pusher.Push(); err != nil {
		prol.Errorf("failed to make initial push to push gateway: %s", err.Error())
	}
	for {
		time.Sleep(time.Second * time.Duration(waitDuration))
		if err := pusher.Push(); err != nil {
			prol.Errorf("failed to push data was collected to push gateway: %s\n", err.Error())
		} else {
			prol.Debugln("pushing data was collected to push gateway")
		}
	}
}
