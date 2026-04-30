###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.26.2 AS build-stage

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
FROM docker.io/library/alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11 AS final-stage

WORKDIR /

COPY --from=build-stage /app/bin/speedtest-exporter /

VOLUME /cache

EXPOSE 8080

USER nobody:nobody

ENTRYPOINT ["/speedtest-exporter"]

#
# END final-stage
###############################################################################
