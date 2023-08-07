# log-generator

![version: 0.6.0](https://img.shields.io/badge/version-0.6.0-informational?style=flat-square) ![type: application](https://img.shields.io/badge/type-application-informational?style=flat-square) ![app version: v0.6.0](https://img.shields.io/badge/app%20version-v0.6.0-informational?style=flat-square) ![kube version: >=1.16.0-0](https://img.shields.io/badge/kube%20version->=1.16.0--0-informational?style=flat-square) [![artifact hub](https://img.shields.io/badge/artifact%20hub-log--generator-informational?style=flat-square)](https://artifacthub.io/packages/helm/kube-logging/log-generator)

A Helm chart for Log-generator

**Homepage:** <https://kube-logging.github.io>

## TL;DR;

```bash
helm repo add kube-logging https://kube-logging.github.io/helm-charts
helm install --generate-name --wait kube-logging/log-generator
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| replicaCount | int | `1` |  |
| image.repository | string | `"ghcr.io/kube-logging/log-generator"` |  |
| image.tag | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `""` |  |
| fullnameOverride | string | `""` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `nil` |  |
| podSecurityContext | object | `{}` |  |
| securityContext | object | `{}` |  |
| app.minInterval | int | `100` |  |
| app.maxInterval | int | `1` |  |
| app.count | int | `-1` |  |
| app.randomise | bool | `true` |  |
| app.eventPerSec | int | `1` |  |
| app.bytePerSec | int | `0` |  |
| app.golang | bool | `false` |  |
| app.nginx | bool | `true` |  |
| app.apache | bool | `false` |  |
| api.addr | string | `"11000"` |  |
| api.serviceName | string | `"log-generator-api"` |  |
| api.basePath | string | `"/"` |  |
| api.serviceMonitor.enabled | bool | `false` |  |
| api.serviceMonitor.additionalLabels | object | `{}` |  |
| api.serviceMonitor.namespace | string | `nil` |  |
| api.serviceMonitor.interval | string | `nil` |  |
| api.serviceMonitor.scrapeTimeout | string | `nil` |  |
| resources | object | `{}` |  |
| nodeSelector | object | `{}` |  |
| tolerations | list | `[]` |  |
| affinity | object | `{}` |  |
