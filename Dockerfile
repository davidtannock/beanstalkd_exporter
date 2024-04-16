FROM golang:1.22-alpine3.19 as build

WORKDIR /go/src/github.com/davidtannock/beanstalkd_exporter

COPY . .

RUN go install -v

FROM alpine:3.19

WORKDIR /
COPY --from=build /go/bin/beanstalkd_exporter /beanstalkd_exporter

ENTRYPOINT ["/beanstalkd_exporter"]
