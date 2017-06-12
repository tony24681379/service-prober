ARG GO_VERSION=1.8.1
FROM golang:${GO_VERSION}-alpine AS build-stage
WORKDIR /go/src/github.com/tony24681379/service-prober
COPY ./ /go/src/github.com/tony24681379/service-prober
RUN go test $(go list ./... | grep -v /vendor/) \
  && go install

FROM alpine:3.5
COPY --from=build-stage /go/bin/service-prober .
ENTRYPOINT ["/service-prober"]