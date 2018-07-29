package server

import (
	"net/http"

	"github.com/davidtannock/beanstalkd_exporter/pkg/beanstalkd"
	"github.com/davidtannock/beanstalkd_exporter/pkg/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

// Opts contains the options for configuring the http server.
type Opts struct {
	ListenAddress string
	MetricsPath   string

	BeanstalkdAddress       string
	BeanstalkdSystemMetrics []string
	BeanstalkdTubes         []string
	BeanstalkdTubeMetrics   []string
}

// ListenAndServe initialises a http server and starts listening
// for http requests.
func ListenAndServe(opts Opts) {
	beanstalkd := beanstalkd.NewServer(opts.BeanstalkdAddress)

	collector := exporter.NewBeanstalkdCollector(beanstalkd, exporter.CollectorOpts{
		SystemMetrics: opts.BeanstalkdSystemMetrics,
		Tubes:         opts.BeanstalkdTubes,
		TubeMetrics:   opts.BeanstalkdTubeMetrics,
	})
	prometheus.MustRegister(collector)

	log.Infoln("Listening on", opts.ListenAddress)
	http.HandleFunc("/", index)
	http.Handle(opts.MetricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(opts.ListenAddress, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html>
	<head>
		<title>Beanstalkd Exporter</title>
	</head>
	<body>
		<h1>Beanstalkd Exporter</h1>
		<p><a href='/metrics'>Metrics</a></p>
	</body>
</html>`))
}
