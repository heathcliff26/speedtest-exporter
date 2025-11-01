###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.1 AS build-stage

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
FROM docker.io/library/alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412 AS final-stage

WORKDIR /

COPY --from=build-stage /app/bin/speedtest-exporter /

EXPOSE 8080

USER 1001

ENTRYPOINT ["/speedtest-exporter"]

#
# END final-stage
###############################################################################
