# speedtest-exporter Helm Chart

This Helm chart deploys the speedtest-exporter - monitor the network speed via prometheus

## Prerequisites

- Kubernetes 1.32+
- Helm 3.19+
- FluxCD installed in the cluster (recommended)

## Installation

### Installing from OCI Registry (GitHub Packages)

```bash
# Install the chart
helm install speedtest-exporter oci://ghcr.io/heathcliff26/manifests/speedtest-exporter --version <version>
```

## Configuration

### Minimal Configuration (No Ingress)

For local development or testing use the default values.

## Values Reference

See [values.yaml](./values.yaml) for all available configuration options.

### Key Parameters

| Parameter                | Description                                          | Default                                   |
| ------------------------ | ---------------------------------------------------- | ----------------------------------------- |
| `image.repository`       | Container image repository                           | `ghcr.io/heathcliff26/speedtest-exporter` |
| `image.tag`              | Container image tag                                  | Same as chart version                     |
| `replicaCount`           | Number of replicas                                   | `1`                                       |
| `ingress.enabled`        | Enable ingress                                       | `false`                                   |
| `servicemonitor.enabled` | Create a Service Monitor for the prometheus operator | `false`                                   |

## Support

For more information, visit: https://github.com/heathcliff26/speedtest-exporter
