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

package collectors

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	prol "github.com/prometheus/common/log"
	"github.com/prometheus/node_exporter/collector"

	"github.com/bizflycloud/bizfly-agent/client"
)

// NewNodeCollector ...
func NewNodeCollector(collectors []string) (*NodeCollector, error) {
	logger := log.NewLogfmtLogger(os.Stdout)
	c, err := collector.NewNodeCollector(logger, collectors...)
	if err != nil {
		return nil, err
	}

	nc := &NodeCollector{
		collectFunc:  c.Collect,
		describeFunc: c.Describe,
		collectorsFunc: func() map[string]collector.Collector {
			return c.Collectors
		},
		httpClient:    client.NewHTTPClient(),
		deviceMetrics: []string{"node_filesystem_size_bytes", "node_filesystem_free_bytes"},
	}

	return nc, nil
}

// NodeCollector ...
type NodeCollector struct {
	collectFunc    func(ch chan<- prometheus.Metric)
	describeFunc   func(ch chan<- *prometheus.Desc)
	collectorsFunc func() map[string]collector.Collector
	httpClient     *client.Client
	deviceMetrics  []string
}

// Collectors ...
func (n *NodeCollector) Collectors() map[string]collector.Collector {
	return n.collectorsFunc()
}

// Name ...
func (n *NodeCollector) Name() string {
	return "node"
}

// Collect ...
func (n *NodeCollector) Collect(ch chan<- prometheus.Metric) {
	mChan := make(chan prometheus.Metric, 1)
	go func() {
		defer close(mChan)
		n.collectFunc(mChan)
	}()
	for m := range mChan {
		d := strings.ToLower(m.Desc().String())
		if n.IsDeviceMetric(d) {
			ch <- n.metricWithDeviceMappings(m)
		} else {
			ch <- m
		}
	}
}

// IsDeviceMetric ...
func (n *NodeCollector) IsDeviceMetric(desc string) bool {
	for _, s := range n.deviceMetrics {
		if strings.Contains(desc, fmt.Sprintf(`fqname: "%s"`, s)) {
			return true
		}
	}
	return false
}

// Describe ...
func (n *NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	n.describeFunc(ch)
}

var errDeviceNotInMapping = errors.New("device not in mapping")

type deviceMappingMetric struct {
	metric        prometheus.Metric
	n             *NodeCollector
	deviceMapping map[string]string
}

func (m deviceMappingMetric) updateDeviceLabel(label *dto.LabelPair) error {
	mountName := label.GetValue()
	for k, v := range m.deviceMapping {
		if strings.HasPrefix(mountName, k) {
			label.Value = &v
			return nil
		}
	}
	return errDeviceNotInMapping
}

func (m deviceMappingMetric) Desc() *prometheus.Desc { return m.metric.Desc() }

func (m deviceMappingMetric) Write(pb *dto.Metric) error {
	e := m.metric.Write(pb)
	for _, label := range pb.Label {
		if label.GetName() == "device" {
			err := m.updateDeviceLabel(label)
			if err != nil {
				prol.Errorf("Can't update label device %v", label)
			}
			break
		}
	}
	return e
}

func (n *NodeCollector) metricWithDeviceMappings(m prometheus.Metric) prometheus.Metric {
	return deviceMappingMetric{metric: m, n: n, deviceMapping: getDeviceMapping()}
}
