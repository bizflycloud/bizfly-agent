module git.paas.vn/OpenStack-Infra/bizfly-agent

go 1.12

require (
	github.com/prometheus/client_golang v1.6.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/node_exporter v0.18.1
	github.com/spf13/viper v1.7.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/prometheus/node_exporter v0.18.1 => github.com/prometheus/node_exporter v0.18.1-0.20190612184716-2bc133cd486e
