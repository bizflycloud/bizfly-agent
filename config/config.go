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

// Configurations ...
type Configurations struct {
	Agent      AgentsConfigurations
	AuthServer ServersConfigurations
	PushGW     PushGateWay
	ConfigDir  string
}

// AgentsConfigurations ...
type AgentsConfigurations struct {
	ID       string
	Name     string
	Hostname string
}

// ServersConfigurations ...
type ServersConfigurations struct {
	DefaultMetadataEndpoint string
}

// PushGateWay ...
type PushGateWay struct {
	URL          string
	WaitDuration int
}
