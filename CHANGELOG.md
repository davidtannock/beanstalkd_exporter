# Changelog

## 2.0.0 / 2024-04-16

* [CHANGE] Nix!
    * https://github.com/nix-community/gomod2nix
    * nix develop
    * nix build
* [CHANGE] Upgraded to go 1.22.1
* [CHANGE] Migrated from dep (https://github.com/golang/dep) to go modules
* [CHANGE] Removed vendor dependencies
* [CHANGE] Removed github.com/prometheus/common dependency (considered internal to prometheus projects)
* [CHANGE] Changed logger to log/slog, and reduced the number of log lines
* [CHANGE] Restructured code (pkg => internal)
* [CHANGE] Makefile changes:
    * Commands mostly unchanged
    * Kept for those who don't want to use nix
* [CHANGE] Switched to github.com/urfave/cli/v2 for cli command
* [CHANGE] Cli flag changes:
    * Removed logging flags `log.level` and `log.format`
    * Added flag `beanstalkd.dialTimeout`
    * Added flag `beanstalkd.keepAlivePeriod`

## 0.3.1 / 2020-07-13

* [ENHANCEMENT] Report test coverage
* [BUGFIX] Fixed yml formatting
* [BUGFIX] Fixed staticcheck

## 0.3.0 / 2019-02-23

* [ENHANCEMENT] Cleaning up code to improve the code climate score

## 0.2.0 / 2019-02-21

* [CHANGE] Added flag `beanstalkd.allTubes` to collect metrics for all tubes
* [ENHANCEMENT] README.md improvements
* [BUGFIX] Fixed Travis CI test failures

## 0.1.1 / 2018-08-05

* [CHANGE] Export the exporter.BeanstalkdServer interface
* [ENHANCEMENT] Travis CI
* [ENHANCEMENT] Better README.md
* [BUGFIX] Updated .dockerignore

## 0.1.0 / 2018-08-04

* [ENHANCEMENT] Added many tests, refactoring code along the way (no feature changes)
* [BUGFIX] Fixed the Makefile docker command

## 0.0.2 / 2018-07-29

* [BUGFIX] Fixed the help text of the `cmd-stats-job` stat

## 0.0.1 / 2018-07-29

* [FEATURE] Initial commit of the project
