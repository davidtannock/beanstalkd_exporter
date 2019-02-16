package server

import (
	"net/http"

	"github.com/davidtannock/beanstalkd_exporter/pkg/beanstalkd"
	"github.com/davidtannock/beanstalkd_exporter/pkg/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

var (
	metricsPath string
)

// Opts contains the options for configuring the http server.
type Opts struct {
	ListenAddress string
	MetricsPath   string

	BeanstalkdAddress       string
	BeanstalkdSystemMetrics []string
	BeanstalkdAllTubes      bool
	BeanstalkdTubes         []string
	BeanstalkdTubeMetrics   []string
}

// ListenAndServe initialises a http server and starts listening
// for http requests.
func ListenAndServe(opts Opts, logger log.Logger) {
	metricsPath = opts.MetricsPath

	// Fetching all tubes overrides specific tubes.
	tubes := opts.BeanstalkdTubes
	if opts.BeanstalkdAllTubes {
		tubes = nil
	}

	collector, err := exporter.NewBeanstalkdCollector(
		beanstalkd.NewServer(opts.BeanstalkdAddress),
		exporter.CollectorOpts{
			SystemMetrics: opts.BeanstalkdSystemMetrics,
			AllTubes:      opts.BeanstalkdAllTubes,
			Tubes:         tubes,
			TubeMetrics:   opts.BeanstalkdTubeMetrics,
		},
		logger,
	)
	if err != nil {
		logger.Fatal(err)
	}

	prometheus.MustRegister(collector)

	logger.Infoln("Listening on", opts.ListenAddress)
	http.HandleFunc("/", index)
	http.Handle(opts.MetricsPath, promhttp.Handler())
	logger.Fatal(http.ListenAndServe(opts.ListenAddress, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html>
	<head>
		<title>Beanstalkd Exporter</title>
	</head>
	<body>
		<h1>Beanstalkd Exporter</h1>
		<p><a href='` + metricsPath + `'>Metrics</a></p>
	</body>
</html>`))
}
