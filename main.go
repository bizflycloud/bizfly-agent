package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/log"
	"github.com/prometheus/node_exporter/collector"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var pushGatewayAddress = kingpin.Flag("pushgateway.address", "The address of pushgateway server").Default("http://localhost/metrics").String()
	var waitDuration = kingpin.Flag("wait.duration", "Time in seconds to wait before pushing to push gateway").Default("30").Int()

	log.AddFlags(kingpin.CommandLine)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	httpClient := &client{
		httpClient:       http.DefaultClient,
		metadataEndpoint: defaultMetadataEndpoint,
	}

	nc, err := collector.NewNodeCollector(defaultCollectors...)
	if err != nil {
		log.Fatalf("failed to create new collector: %s\n", err.Error())
	}
	data, err := ioutil.ReadFile("/var/lib/cloud/data/instance-id")
	if err != nil {
		log.Fatalf("failed to read instance id file: %s\\n", err.Error())
	}
	instanceID := strings.TrimSuffix(string(data), "\n")

	pusher := push.New(*pushGatewayAddress, "bizfly-agent").
		Client(httpClient).
		Grouping("instance_id", instanceID).
		Collector(nc)

	if err := pusher.Push(); err != nil {
		log.Errorf("failed to make initial push to push gateway: %s\n", err.Error())
	}
	for {
		time.Sleep(time.Second * time.Duration(*waitDuration))
		if err := pusher.Push(); err != nil {
			log.Errorf("failed to push to push gateway: %s\n", err.Error())
		}
	}
}
