module github.com/bizflycloud/bizfly-agent

go 1.14

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/go-kit/kit v0.10.0
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/prometheus/client_golang v1.6.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.10.0
	github.com/prometheus/node_exporter v1.0.0
	github.com/shirou/gopsutil v2.20.6+incompatible
	github.com/spf13/viper v1.7.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/prometheus/node_exporter => github.com/bizflycloud/node_exporter v1.0.2
