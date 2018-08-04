package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/davidtannock/beanstalkd_exporter/pkg/server"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

type options struct {
	listenAddress           *string
	metricsPath             *string
	beanstalkdAddress       *string
	beanstalkdSystemMetrics *string
	beanstalkdTubes         *string
	beanstalkdTubeMetrics   *string
}

func initApplication(app *kingpin.Application) *options {
	opts := &options{}

	opts.listenAddress = app.Flag(
		"web.listen-address",
		"Address to listen on for web interface and telemetry.",
	).Default(":8080").String()

	opts.metricsPath = app.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()

	opts.beanstalkdAddress = app.Flag(
		"beanstalkd.address",
		"Address of beanstalkd instance.",
	).Default("localhost:11300").String()

	opts.beanstalkdSystemMetrics = app.Flag(
		"beanstalkd.systemMetrics",
		"Comma separated beanstalkd system metrics to collect. All metrics will be collected if this is not set.",
	).Default("").String()

	opts.beanstalkdTubes = app.Flag(
		"beanstalkd.tubes",
		"Comma separated beanstalkd tubes for which to collect metrics. No tube metrics will be collected if this is not set.",
	).Default("").String()

	opts.beanstalkdTubeMetrics = app.Flag(
		"beanstalkd.tubeMetrics",
		"Comma separated beanstalkd tube metrics to collect for the specified tubes. All metrics will be collected if this is not set.",
	).Default("").String()

	log.AddFlags(app)

	return opts
}

func main() {
	logger := log.Base()
	app := kingpin.New(filepath.Base(os.Args[0]), "")

	opts := initApplication(app)
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		app.Fatalf("%s, try --help", err)
	}

	logger.Infof("Starting beanstalkd_exporter")

	sOpts := server.Opts{
		ListenAddress:           *opts.listenAddress,
		MetricsPath:             *opts.metricsPath,
		BeanstalkdAddress:       *opts.beanstalkdAddress,
		BeanstalkdSystemMetrics: toStringArray(*opts.beanstalkdSystemMetrics),
		BeanstalkdTubes:         toStringArray(*opts.beanstalkdTubes),
		BeanstalkdTubeMetrics:   toStringArray(*opts.beanstalkdTubeMetrics),
	}
	server.ListenAndServe(sOpts, logger)
}

func toStringArray(flag string) []string {
	flags := []string{}
	for _, part := range strings.Split(flag, ",") {
		if s := strings.Trim(part, " "); s != "" {
			flags = append(flags, s)
		}
	}
	return flags
}
