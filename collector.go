package main

import (
	"encoding/json"
	"errors"
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

	nc := &nodeCollector{
		collectFunc:  c.Collect,
		describeFunc: c.Describe,
		collectorsFunc: func() map[string]collector.Collector {
			return c.Collectors
		},
		httpClient:    newHTTPClient(),
		deviceMetrics: []string{"node_filesystem_size_bytes", "node_filesystem_free_bytes"},
		deviceUsed:    map[string]struct{}{},
	}
	if data, err := nc.httpClient.DeviceMapping(); err == nil {
		json.Unmarshal(data, &nc.deviceMapping)
	}

	return nc, nil
}

type nodeCollector struct {
	collectFunc    func(ch chan<- prometheus.Metric)
	describeFunc   func(ch chan<- *prometheus.Desc)
	collectorsFunc func() map[string]collector.Collector
	httpClient     *client
	deviceMetrics  []string
	deviceMapping  map[string]string
	deviceUsed     map[string]struct{}
}

func (n *nodeCollector) updateDeviceMapping() {
	for k := range n.deviceMapping {
		delete(n.deviceMapping, k)
	}
	if data, err := n.httpClient.DeviceMapping(); err == nil {
		json.Unmarshal(data, &n.deviceMapping)
	}
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

func (n *nodeCollector) cleanUnusedDevice() {
	for device := range n.deviceMapping {
		if _, ok := n.deviceUsed[device]; !ok {
			delete(n.deviceMapping, device)
		}
	}
}

func (n *nodeCollector) Describe(ch chan<- *prometheus.Desc) {
	n.describeFunc(ch)
}

var errDeviceNotInMapping = errors.New("device not in mapping")

type deviceMappingMetric struct {
	metric prometheus.Metric
	n      *nodeCollector
}

func (m deviceMappingMetric) updateDeviceLabel(label *dto.LabelPair) error {
	mountName := label.GetValue()
	for k, v := range m.n.deviceMapping {
		if strings.HasPrefix(mountName, k) {
			label.Value = &v
			m.n.deviceUsed[k] = struct{}{}
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
			// update with current mapping first
			if err := m.updateDeviceLabel(label); err != nil {
				// the device mapping has change, update and recheck
				m.n.updateDeviceMapping()
				m.updateDeviceLabel(label)
			}
			break
		}
	}
	return e
}

func (n *nodeCollector) metricWithDeviceMappings(m prometheus.Metric) prometheus.Metric {
	return deviceMappingMetric{metric: m, n: n}
}
