###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27.1 AS build-stage

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
FROM docker.io/library/alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS final-stage

WORKDIR /

COPY --from=build-stage /app/bin/speedtest-exporter /

VOLUME /cache

EXPOSE 8080

USER nobody:nobody

ENTRYPOINT ["/speedtest-exporter"]

#
# END final-stage
###############################################################################
