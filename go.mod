module git.paas.vn/OpenStack-Infra/bizfly-agent

go 1.12

require (
	github.com/prometheus/client_golang v0.9.4
	github.com/prometheus/common v0.4.1
	github.com/prometheus/node_exporter v0.18.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/prometheus/client_golang v0.9.4 => github.com/prometheus/client_golang v0.9.3

replace github.com/prometheus/procfs v0.0.2 => github.com/prometheus/procfs v0.0.0-20190529155944-65bdadfa96ae
