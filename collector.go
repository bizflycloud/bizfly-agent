package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/node_exporter/collector"
)

var defaultCollectors = []string{
	"cpu",
	"diskstats",
	"filesystem",
	"loadavg",
	"meminfo",
	"netstat",
	"netdev",
}

func newNodeCollector(collectors []string) (*nodeCollector, error) {
	c, err := collector.NewNodeCollector(collectors...)
	if err != nil {
		return nil, err
	}

	return &nodeCollector{
		collectFunc:  c.Collect,
		describeFunc: c.Describe,
		collectorsFunc: func() map[string]collector.Collector {
			return c.Collectors
		},
		httpClient:    newHTTPClient(),
		deviceMetrics: []string{"node_filesystem_size_bytes", "node_filesystem_free_bytes"},
	}, nil
}

type nodeCollector struct {
	collectFunc    func(ch chan<- prometheus.Metric)
	describeFunc   func(ch chan<- *prometheus.Desc)
	collectorsFunc func() map[string]collector.Collector
	httpClient     *client
	deviceMetrics  []string
}

func (n *nodeCollector) Collectors() map[string]collector.Collector {
	return n.collectorsFunc()
}

func (n *nodeCollector) Name() string {
	return "node"
}

func (n *nodeCollector) Collect(ch chan<- prometheus.Metric) {
	mChan := make(chan prometheus.Metric, 1)
	go func() {
		defer close(mChan)
		n.collectFunc(mChan)
	}()
	for m := range mChan {
		d := strings.ToLower(m.Desc().String())
		if n.isDeviceMetric(d) {
			ch <- n.metricWithDeviceMappings(m)
		} else {
			ch <- m
		}
	}
}

func (n *nodeCollector) isDeviceMetric(desc string) bool {
	for _, s := range n.deviceMetrics {
		if strings.Contains(desc, fmt.Sprintf(`fqname: "%s"`, s)) {
			return true
		}
	}
	return false
}

func (n *nodeCollector) Describe(ch chan<- *prometheus.Desc) {
	n.describeFunc(ch)
}

type deviceMappingMetric struct {
	metric        prometheus.Metric
	deviceMapping map[string]string
}

func (m deviceMappingMetric) Desc() *prometheus.Desc { return m.metric.Desc() }

func (m deviceMappingMetric) Write(pb *dto.Metric) error {
	e := m.metric.Write(pb)
	for _, label := range pb.Label {
		mountName := label.GetValue()
		for k, v := range m.deviceMapping {
			if strings.HasPrefix(mountName, k) {
				label.Value = &v
				break
			}
		}
	}
	return e
}

func (n *nodeCollector) metricWithDeviceMappings(m prometheus.Metric) prometheus.Metric {
	if data, err := n.httpClient.GetDeviceMapping(); err == nil {
		deviceMapping := make(map[string]string)
		if err := json.Unmarshal(data, &deviceMapping); err == nil {
			return deviceMappingMetric{metric: m, deviceMapping: deviceMapping}
		}
	}
	return m
}
