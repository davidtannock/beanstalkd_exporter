FROM golang:1.10-alpine3.8 as build

WORKDIR /go/src/github.com/davidtannock/beanstalkd_exporter

COPY . .

RUN go install -v

FROM alpine:3.8

WORKDIR /
COPY --from=build /go/bin/beanstalkd_exporter /beanstalkd_exporter

ENTRYPOINT ["/beanstalkd_exporter"]