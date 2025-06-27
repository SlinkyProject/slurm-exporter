# Slurm Cluster Metrics Exporter

<div align="center">

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=for-the-badge)](./LICENSES/Apache-2.0.txt)
[![Tag](https://img.shields.io/github/v/tag/SlinkyProject/slurm-exporter?style=for-the-badge)](https://github.com/SlinkyProject/slurm-exporter/tags/)
[![Go-Version](https://img.shields.io/github/go-mod/go-version/SlinkyProject/slurm-exporter?style=for-the-badge)](./go.mod)
[![Last-Commit](https://img.shields.io/github/last-commit/SlinkyProject/slurm-exporter?style=for-the-badge)](https://github.com/SlinkyProject/slurm-exporter/commits/)

</div>

[Prometheus] collector and exporter of [Slurm] cluster metrics. A [Slinky]
project.

## Table of Contents

<!-- mdformat-toc start --slug=github --no-anchors --maxlevel=6 --minlevel=1 -->

- [Slurm Cluster Metrics Exporter](#slurm-cluster-metrics-exporter)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
    - [Nodes](#nodes)
    - [Partitions](#partitions)
    - [User Statistics](#user-statistics)
  - [Limitations](#limitations)
  - [Installation](#installation)
  - [License](#license)

<!-- mdformat-toc end -->

## Overview

Slurm metrics are collected from Slurm through the
[Slurm REST API][slurm-restapi] and exported as Prometheus types. The exporter
is granted authorization via an authentication JWT token, hence it runs with
privileges of a Slurm cluster user via token.

The recommended deployment mechanism is through [Helm] chart.

## Features

### Nodes

- [**Allocated**][node-allocated]: nodes which have been allocated to one or
  more jobs.
- [**Completing**][node-completing]: all jobs associated with this node are in
  the process of COMPLETING.
- [**Down**][node-down]: nodes which are unavailable for use.
- [**Drain**][node-drain]: nodes which are marked as drain, unavailable for
  future work but can complete their current work.
  - [**Drained**][node-drained]: nodes which are unavailable for use (per system
    administrator request) and have no more work to complete.
  - [**Draining**][node-draining]: The node is currently allocated to job(s),
    but will not be allocated additional jobs.
- [**Idle**][node-idle]: nodes which are not allocated to any jobs and is
  available for use.
- [**Maintenance**][node-maint]: nodes which are currently in a *maintenance*
  reservation.
- [**Mixed**][node-mixed]: nodes which have some but not all of their CPUs
  ALLOCATED, or suspended jobs have TRES (e.g. Memory) still allocated.
- [**Reserved**][node-reserved]: nodes which are in an advanced reservation and
  not generally available.

### Partitions

- **Nodes**: number of nodes associated with the partition.
- **CPUs**: number of CPUs associated with the partition.
- **Jobs**: number of incomplete (e.g. pending, running) jobs in the partition.
- **Allocated CPUs**: number of ALLOCATED CPUs among all nodes in the partition.
- **Idle CPUs**: number of IDLE CPUs among all nodes in the partition.
- **Pending Jobs**: number of pending jobs in the partition.
- **Pending Jobs, Max Nodes**: max number of nodes requested among all pending
  jobs in the partition.
- **Running Jobs**: number of running jobs in the partition.
- **Held Jobs**: number of held jobs in the partition.

### User Statistics

- **Job Count**: number of incomplete (e.g. pending, running) jobs for the user.
- [**Pending Jobs**][job-states]: number of pending jobs for the user.
- [**Running Jobs**][job-states]: number of running jobs for the user.
- [**Held Jobs**][job-states]: number of held jobs for the user.

## Limitations

Currently only a minimal set of metrics are collected. More metrics may be added
in the future.

- **Slurm Version**: >=
  [25.05](https://www.schedmd.com/slurm-version-25-05-0-is-now-available/)

## Installation

Install `kube-prometheus-stack` for metrics collection and observation.
Prometheus is also used as an extension API server so custom Slurm metrics may
be used with autoscaling.

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace=prometheus --create-namespace
```

Install the slurm-exporter:

```bash
helm install slurm-exporter oci://ghcr.io/slinkyproject/charts/slurm-exporter \
  --namespace=slurm-exporter --create-namespace
```

## License

Copyright (C) SchedMD LLC.

Licensed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0) you
may not use project except in compliance with the license.

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.

<!-- links -->

[helm]: https://helm.sh/
[job-states]: https://slurm.schedmd.com/job_state_codes.html#states
[node-allocated]: https://slurm.schedmd.com/sinfo.html#OPT_ALLOCATED
[node-completing]: https://slurm.schedmd.com/sinfo.html#OPT_COMPLETING
[node-down]: https://slurm.schedmd.com/sinfo.html#OPT_DOWN
[node-drain]: https://slurm.schedmd.com/scontrol.html#OPT_DRAIN
[node-drained]: https://slurm.schedmd.com/sinfo.html#OPT_DRAINED
[node-draining]: https://slurm.schedmd.com/sinfo.html#OPT_DRAINING
[node-idle]: https://slurm.schedmd.com/sinfo.html#OPT_IDLE
[node-maint]: https://slurm.schedmd.com/sinfo.html#OPT_MAINT
[node-mixed]: https://slurm.schedmd.com/sinfo.html#OPT_MIXED
[node-reserved]: https://slurm.schedmd.com/sinfo.html#OPT_RESERVED
[prometheus]: https://prometheus.io/
[slinky]: https://slinky.ai/
[slurm]: https://slurm.schedmd.com/overview.html
[slurm-restapi]: https://slurm.schedmd.com/rest_api.html
