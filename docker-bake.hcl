// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

################################################################################

variable "REGISTRY" {
  default = "ghcr.io/slinkyproject"
}

variable "VERSION" {
  default = "0.0.0"
}

function "format_tag" {
  params = [registry, stage, version]
  result = format("%s:%s", join("/", compact([registry, stage])), join("-", compact([version])))
}

################################################################################

group "default" {
  targets = [
    "exporter",
  ]
}

target "exporter" {
  dockerfile = "Dockerfile"
  labels = {
    # Ref: https://github.com/opencontainers/image-spec/blob/v1.0/annotations.md
    "org.opencontainers.image.authors" = "slinky@schedmd.com"
    "org.opencontainers.image.documentation" = "https://github.com/SlinkyProject/slurm-exporter"
    "org.opencontainers.image.license" = "Apache-2.0"
    "org.opencontainers.image.vendor" = "SchedMD LLC."
    "org.opencontainers.image.version" = "${VERSION}"
    "org.opencontainers.image.source" = "https://github.com/SlinkyProject/slurm-exporter"
    "org.opencontainers.image.title" = "Slurm Exporter"
    "org.opencontainers.image.description" = "Prometheus collector and exporter for metrics extracted from Slurm"
    # Ref: https://docs.redhat.com/en/documentation/red_hat_software_certification/2025/html/red_hat_openshift_software_certification_policy_guide/assembly-requirements-for-container-images_openshift-sw-cert-policy-introduction#con-image-metadata-requirements_openshift-sw-cert-policy-container-images
    "vendor" = "SchedMD LLC."
    "version" = "${VERSION}"
    "release" = "https://github.com/SlinkyProject/slurm-exporter"
    "name" = "Slurm Exporter"
    "summary" = "Prometheus collector and exporter for metrics extracted from Slurm"
    "description" = "Prometheus collector and exporter for metrics extracted from Slurm"
  }
  tags = [
    format_tag("${REGISTRY}", "slurm-exporter", "${VERSION}"),
  ]
}
