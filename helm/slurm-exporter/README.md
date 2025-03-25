# slurm-exporter

![Version: 0.2.1](https://img.shields.io/badge/Version-0.2.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 24.11](https://img.shields.io/badge/AppVersion-24.11-informational?style=flat-square)

Helm Chart for Slurm Prometheus Exporter

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| exporter.affinity | object | `{}` |  Set affinity for Kubernetes Pod scheduling. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity |
| exporter.enabled | bool | `true` |  Enables metrics collection. |
| exporter.image.repository | string | `"ghcr.io/slinkyproject/slurm-exporter"` |  Set the image repository to use. |
| exporter.image.tag | string | The chart Version. |  Set the image tag to use. |
| exporter.imagePullPolicy | string | `"IfNotPresent"` |  Set the image pull policy. |
| exporter.priorityClassName | string | `""` |  Set the priority class to use. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass |
| exporter.replicas | integer | `1` |  Set the number of replicas to deploy. |
| exporter.resources | object | `{}` |  Set container resource requests and limits for Kubernetes Pod scheduling. Ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-requests-and-limits-of-pod-and-container |
| exporter.secretName | string | `""` |  |
| exporter.serviceMonitor.enabled | bool | `true` |  |
| exporter.serviceMonitor.endpoints[0].interval | string | `"10s"` |  |
| exporter.serviceMonitor.endpoints[0].path | string | `"/metrics"` |  |
| exporter.serviceMonitor.endpoints[0].port | string | `"metrics"` |  |
| exporter.serviceMonitor.endpoints[0].scheme | string | `"http"` |  |
| grafana.enabled | bool | `true` |  Enables grafana dashboard. |
| imagePullSecrets | list | `[]` |  Set the secrets for image pull. Ref: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/ |
| nameOverride | string | `""` |  Overrides the name of the release. |
| namespaceOverride | string | `""` |  Overrides the namespace of the release. |
| priorityClassName | string | `""` |  Set the priority class to use. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass |

