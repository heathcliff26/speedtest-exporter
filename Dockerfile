###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.21.6@sha256:5f5d61dcb58900bc57b230431b6367c900f9982b583adcabf9fa93fd0aa5544a AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH

WORKDIR /app

COPY vendor ./vendor
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH}" go build -ldflags="-w -s" -o /speedtest-exporter ./cmd/

#
# END build-stage
###############################################################################

###############################################################################
# BEGIN test-stage
# Run the tests in the container
FROM docker.io/library/golang:1.21.6@sha256:5f5d61dcb58900bc57b230431b6367c900f9982b583adcabf9fa93fd0aa5544a AS test-stage

WORKDIR /app

COPY --from=build-stage /app /app
# Not needed for testing, but needed for later stage
COPY --from=build-stage /speedtest-exporter /

RUN go test -v ./...

#
# END test-stage
###############################################################################

###############################################################################
# BEGIN combine-stage
# Combine all outputs, to enable single layer copy for the final image
FROM scratch AS combine-stage

COPY --from=test-stage /speedtest-exporter /

COPY --from=test-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

#
# END combine-stage
###############################################################################

###############################################################################
# BEGIN final-stage
# Create final docker image
FROM scratch AS final-stage

WORKDIR /

COPY --from=combine-stage / /

EXPOSE 8080

USER 1001

ENTRYPOINT ["/speedtest-exporter"]

#
# END final-stage
###############################################################################
