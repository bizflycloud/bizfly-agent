module github.com/bizflycloud/bizfly-agent

go 1.14

require (
	github.com/go-kit/kit v0.10.0
	github.com/mindprince/gonvml v0.0.0-20190828220739-9ebdce4bb989 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.15.0
	github.com/prometheus/node_exporter v1.0.1
	github.com/shirou/gopsutil v3.20.10+incompatible
	github.com/spf13/viper v1.7.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/prometheus/node_exporter => github.com/bizflycloud/node_exporter v1.0.6
