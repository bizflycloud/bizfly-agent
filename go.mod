module git.paas.vn/OpenStack-Infra/bizfly-agent

go 1.12

require (
	github.com/prometheus/client_golang v0.9.4
	github.com/prometheus/common v0.4.1
	github.com/prometheus/node_exporter v0.18.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/prometheus/node_exporter v0.18.1 => github.com/prometheus/node_exporter v0.18.1-0.20190612184716-2bc133cd486e
