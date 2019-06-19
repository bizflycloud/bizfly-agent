module git.paas.vn/OpenStack-Infra/bizfly-agent

go 1.12

require (
	github.com/prometheus/client_golang v0.9.4
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.4.1
	github.com/prometheus/node_exporter v0.18.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/prometheus/node_exporter v0.18.1 => github.com/prometheus/node_exporter v0.18.1-0.20190612184716-2bc133cd486e
