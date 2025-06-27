# Development

## Table of Contents

<!-- mdformat-toc start --slug=github --no-anchors --maxlevel=6 --minlevel=1 -->

- [Development](#development)
  - [Table of Contents](#table-of-contents)
  - [Build](#build)
  - [Deploy](#deploy)

<!-- mdformat-toc end -->

## Build

To build, you need to install build dependencies. To do this, install
[golang](https://go.dev/doc/install) and make.

To build the OCI container:

```sh
make docker-bake
```

## Deploy

```sh
cd helm/slurm-exporter
skaffold run -p dev
```
