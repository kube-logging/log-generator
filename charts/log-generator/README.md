# log-generator

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
