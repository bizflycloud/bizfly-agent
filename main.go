package main

import (
	"net/http"
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
		panic(err)
	}

	pusher := push.New(*pushGatewayAddress, "bizfly-agent").Client(httpClient)
	pusher.Collector(nc)

	if err := pusher.Push(); err != nil {
		log.Errorf("failed to make initial push to push gateway: %v", err)
	}
	for {
		time.Sleep(time.Second * time.Duration(*waitDuration))
		if err := pusher.Push(); err != nil {
			log.Errorf("failed to push to push gateway: %v", err)
		}
	}
}
