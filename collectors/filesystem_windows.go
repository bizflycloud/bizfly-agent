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

// +build windows

package collectors

import (
	"github.com/shirou/gopsutil/disk"
)

func getDeviceMapping() map[string]string {
	m := make(map[string]string)
	partitions, _ := disk.Partitions(false)
	for _, partition := range partitions {
		m[partition.Mountpoint] = partition.Device
	}
	return m
}
