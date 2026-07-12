###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.26.4 AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH

WORKDIR /app

COPY . ./

RUN GOOS=linux GOARCH="${TARGETARCH}" hack/build.sh

#
# END build-stage
###############################################################################

###############################################################################
# BEGIN final-stage
# Create final docker image
FROM docker.io/library/alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b AS final-stage

WORKDIR /

COPY --from=build-stage /app/bin/speedtest-exporter /

VOLUME /cache

EXPOSE 8080

USER nobody:nobody

ENTRYPOINT ["/speedtest-exporter"]

#
# END final-stage
###############################################################################
