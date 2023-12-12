<!-- markdownlint-disable MD033 -->
<h1 align="center">
  <img src="assets/logo.png" alt="irme logo" width="150" height="150" style="border-radius: 25%">
</h1>

<h4 align="center">image-registry-metrics-exporter - Monitor your OCI registry</h4>

<div align="center">
  <a href="https://github.com/radiofrance/image-registry-metrics-exporter/issues/new">Report a Bug</a> ·
  <a href="https://github.com/radiofrance/image-registry-metrics-exporter/issues/new">Request a Feature</a> ·
  <a href="https://github.com/radiofrance/image-registry-metrics-exporter/discussions">Ask a Question</a>
  <br/>
  <br/>

[![GoReportCard](https://goreportcard.com/badge/github.com/radiofrance/image-registry-metrics-exporter)](https://goreportcard.com/report/github.com/radiofrance/image-registry-metrics-exporter)
[![Codecov branch](https://img.shields.io/codecov/c/github/radiofrance/image-registry-metrics-exporter/main?label=code%20coverage)](https://app.codecov.io/gh/radiofrance/image-registry-metrics-exporter/tree/main)
[![GoDoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/radiofrance/image-registry-metrics-exporter)
<br/>
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/radiofrance/image-registry-metrics-exporter?logo=go&logoColor=white&logoWidth=20)
[![License](https://img.shields.io/badge/license-CeCILL%202.1-blue?logo=git&logoColor=white&logoWidth=20)](LICENSE)

<a href="#about">About</a> ·
<a href="#install">How to Install?</a> ·
<a href="#exported-metrics">Metrics</a> ·
<a href="#support">Support</a> ·
<a href="#contributing">Contributing</a> ·
<a href="#security">Security</a>

</div>

---
<!-- markdownlint-enable MD033 -->

## About

**Image Registry Metrics Exporter (IRME)** provides metrics about creation and upload time of your OCI images on 
a compatible registry.

At this time, we support the following registries:
- [Google Container Registry](https://cloud.google.com/container-registry) (gcr.io)

## Install

### Go

```shell
go install github.com/radiofrance/image-registry-metrics-exporter/cmd/image-registry-metrics-exporter@latest
image-registry-metrics-exporter ...
```

### Docker

```shell
docker pull ghcr.io/radiofrance/image-registry-metrics-exporter
docker run --publish 8080 ghcr.io/radiofrance/image-registry-metrics-exporter
```

### Helm

```shell
helm repo add radiofrance-irme https://radiofrance.github.io/image-registry-metrics-exporter
helm upgrade --install image-registry-metrics-exporter radiofrance-irme/image-registry-metrics-exporter \
  --namespace image-registry-metrics-exporter \
  --create-namespace \
  --wait
helm test image-registry-metrics-exporter --namespace image-registry-metrics-exporter
```

## Usage

### Configuration

A configuration file is required to run IRME. It defines when and where it should analyze your OCI images.  
This configuration file can be located in `/etc/irme/config.yaml`, `$HOME/.irme/config.yaml` or in your
current directory (`./config.yaml`).

```yaml
cron: <cron>                    # cron expression used to schedule when the registries will be scanned
registries:                     # list all registries to be scanned
  - provider: <provider name>   # defines which provider to use for this registry
    domain: <domain>            # registry entrypoint
    imagesFilters:              # regular expression to use to filter which repositories or images should be scanned
      - <filter>
    tagsFilters:                # same as `imagesFilters` but for tags
      - <filter>
    rateLimitAPI: <rps_limit>   # limit the number of request per seconds on the registry
    maxConcurrentJobs: <limit>  # set the number of images processed in parallel
```

> For exemple, if we want to scan this project every monday at 8:00: 
>
> ```yaml
> cron: 0 8 * * 1
> registries:
>   - provider: github # NOTE: this provider is currently not available
>     domain: ghcr.io
>     imagesFilters:
>       - radiofrance/image-registry-metrics-exporter
>     tagsFilters:
>       - .*
>     rateLimitAPI: 5
>     maxConcurrentJobs: 5
>```

### Google Container Registry authentication

To use _IRME_ with the GCR provider (`google`), you need to [create a service account](https://cloud.google.com/iam/docs/service-accounts-create) 
with following roles:
- `roles/browser`: allow _IRME_ to list all images available on repositories.
- `roles/storage.objectViewer`: allow _IRME_ to get information about the images themselves

> **NOTE**: If you use _IRME_ outside Helm, the credentials should be exported through the environment variable 
> `GOOGLE_APPLICATION_CREDENTIALS` or configured on your local docker configuration file.

## Exported metrics

| Metric name                                 | Description                  | Labels         |
|---------------------------------------------|------------------------------|----------------|
| `image_registry_exporter_tag_build_time`    | Build timestamp of an image  | [image], [tag] |
| `image_registry_exporter_tag_uploaded_time` | Upload timestamp of an image | [image], [tag] |

## Support

Reach out to the maintainer at one of the following places:

- [GitHub Discussions](https://github.com/radiofrance/image-registry-metrics-exporter/discussions)
- Open an issue on [GitHub](https://github.com/radiofrance/image-registry-metrics-exporter/issues/new)

## Contributing

First off, thanks for taking the time to contribute! Contributions are what make the
open-source community such an amazing place to learn, inspire, and create. Any contributions
you make will benefit everybody else and are **greatly appreciated**.

Please read [our contribution guidelines](docs/CONTRIBUTING.md), and thank you for being involved!

## Security

`image-registry-metrics-exporter` follows good practices of security, but 100% security cannot be assured.
`image-registry-metrics-exporter` is provided **"as is"** without any **warranty**. Use at your own risk.

*For more information and to report security issues, please refer to our [security documentation](docs/SECURITY.md).*

## License

This project is licensed under the **CeCILL License 2.1**.

See [LICENSE](LICENSE) for more information.

## Acknowledgements

Thanks for these awesome resources and projects that were used during development:

- <https://github.com/go-co-op/gocron> - A Golang Job Scheduling Package
- <https://github.com/gorilla/mux> - A powerful HTTP router and URL matcher for building Go web servers with 🦍
- <https://github.com/spf13/viper> - Go configuration with fangs
