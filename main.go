package main

import (
	"strings"

	"github.com/davidtannock/beanstalkd_exporter/pkg/server"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress           *string
	metricsPath             *string
	beanstalkdAddress       *string
	beanstalkdSystemMetrics *string
	beanstalkdTubes         *string
	beanstalkdTubeMetrics   *string
)

func init() {
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address to listen on for web interface and telemetry.",
	).Default(":8080").String()

	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()

	beanstalkdAddress = kingpin.Flag(
		"beanstalkd.address",
		"Address of beanstalkd instance.",
	).Default("localhost:11300").String()

	beanstalkdSystemMetrics = kingpin.Flag(
		"beanstalkd.systemMetrics",
		"Comma separated beanstalkd system metrics to collect. All metrics will be collected if this is not set.",
	).Default("").String()

	beanstalkdTubes = kingpin.Flag(
		"beanstalkd.tubes",
		"Comma separated beanstalkd tubes for which to collect metrics. No tube metrics will be collected if this is not set.",
	).Default("").String()

	beanstalkdTubeMetrics = kingpin.Flag(
		"beanstalkd.tubeMetrics",
		"Comma separated beanstalkd tube metrics to collect for the specified tubes. All metrics will be collected if this is not set.",
	).Default("").String()

	log.AddFlags(kingpin.CommandLine)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
}

func main() {
	log.Infof("Starting beanstalkd_exporter")

	server.ListenAndServe(server.Opts{
		ListenAddress:           *listenAddress,
		MetricsPath:             *metricsPath,
		BeanstalkdAddress:       *beanstalkdAddress,
		BeanstalkdSystemMetrics: toStringArray(*beanstalkdSystemMetrics),
		BeanstalkdTubes:         toStringArray(*beanstalkdTubes),
		BeanstalkdTubeMetrics:   toStringArray(*beanstalkdTubeMetrics),
	})
}

func toStringArray(flag string) []string {
	var flags []string
	for _, part := range strings.Split(flag, ",") {
		if s := strings.Trim(part, " "); s != "" {
			flags = append(flags, s)
		}
	}
	return flags
}
