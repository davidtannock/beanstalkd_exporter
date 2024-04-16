[![Go Report Card](https://goreportcard.com/badge/github.com/davidtannock/beanstalkd_exporter)][goreportcard]
[![Maintainability](https://api.codeclimate.com/v1/badges/5653f08f506a6f02d786/maintainability)][codeclimate]
[![Test Coverage](https://api.codeclimate.com/v1/badges/5653f08f506a6f02d786/test_coverage)][codecoverage]
[![Build Status](https://travis-ci.org/davidtannock/beanstalkd_exporter.png?branch=main)][travisci]

[goreportcard]: https://goreportcard.com/report/github.com/davidtannock/beanstalkd_exporter
[codeclimate]: https://codeclimate.com/github/davidtannock/beanstalkd_exporter/maintainability
[codecoverage]: https://codeclimate.com/github/davidtannock/beanstalkd_exporter/test_coverage
[travisci]: https://travis-ci.org/davidtannock/beanstalkd_exporter

# Beanstalkd Exporter for Prometheus

This is a simple server that scrapes [beanstalkd][beanstalkd] stats and exports them via http
for [Prometheus][prometheus] consumption.

[beanstalkd]: http://kr.github.io/beanstalkd/
[prometheus]: https://prometheus.io/

## Getting Started

To run it:

```bash
$ ./beanstalkd_exporter [flags]
```

Help on flags:

```bash
$ ./beanstalkd_exporter --help
```

## Usage

Specify the address of the beanstalkd instance using the `--beanstalkd.address` flag. For example,

```bash
$ ./beanstalkd_exporter --beanstalkd.address=127.0.0.1:11300
```

The default address is `localhost:11300`.

### Example With Beanstalkd

Start a beanstalkd instance with the following docker command.

```bash
$ docker run -d -p 11300:11300 --name beanstalkd dtannock/beanstalkd:latest
```

Run beanstalkd_exporter. The default flag values should be able to connect to the running beanstalkd instance.

```bash
$ ./beanstalkd_exporter
```

Fetch the metrics for Prometheus.

```bash
$ curl -s http://localhost:8080/metrics
```

### Failed Scrapes

Of the failed scraping strategies described [here][failedscrapes], the `up` variable is used.

You can see this in practice by starting/stopping the docker container, and fetching the metrics,
then find the `beanstalkd_up` value.

```bash
$ docker stop beanstalkd
$ curl -s http://localhost:8080/metrics | grep beanstalkd_up
$ docker start beanstalkd
$ curl -s http://localhost:8080/metrics | grep beanstalkd_up
```

[failedscrapes]: https://prometheus.io/docs/instrumenting/writing_exporters/#failed-scrapes

## Metrics

Without passing any flags, only the system-level stats will be collected from beanstalkd
(i.e. tube-level stats will not be collected).

To collect tube-level stats, you must use either the `--beanstalkd.allTubes` or `--beanstalkd.tubes` flag.

`--beanstalkd.allTubes` will collect metrics for all tubes.

```bash
$ ./beanstalkd_exporter --beanstalkd.allTubes
```

`--beanstalkd.tubes` will collect metrics for one or more specific tubes.

```bash
$ ./beanstalkd_exporter --beanstalkd.tubes=default,anotherTube
```

The metrics collected from beanstalkd can be filtered using the `--beanstalkd.systemMetrics` and
`--beanstalkd.tubeMetrics` flags. For example,

```bash
$ ./beanstalkd_exporter \
    --beanstalkd.systemMetrics=current_jobs_urgent_count,current_jobs_ready_count \
    --beanstalkd.tubes=default \
    --beanstalkd.tubeMetrics=tube_current_jobs_ready_count
```

Will fetch only 2 system-level metrics, and 1 metric labelled for the `default` tube.

The full list of metrics is available on [this page][metrics].

[metrics]: https://github.com/davidtannock/beanstalkd_exporter/blob/main/internal/exporter/metrics.go

## Development

### Building

```bash
$ make build
```

### Testing

```bash
$ make test
```

## Version 2

Version 2 was an exercise in learning Nix (https://nixos.org/), specifically:
* Nix development environments
* Building go projects with Nix (thank you https://github.com/nix-community/gomod2nix)
* Nix flakes

The other [changes](https://github.com/davidtannock/beanstalkd_exporter/blob/main/CHANGELOG.md) are mainly related to removing legacy dependencies. The cli command is largely unchanged, and nix is not necessary to build the executable (see [Makefile](https://github.com/davidtannock/beanstalkd_exporter/blob/main/Makefile)).

## License

Apache License 2.0, see [LICENSE](https://github.com/davidtannock/beanstalkd_exporter/blob/main/LICENSE).
