# OpenHue CLI
[![Build](https://github.com/openhue/openhue-cli/actions/workflows/build.yml/badge.svg)](https://github.com/openhue/openhue-cli/actions/workflows/build.yml)
[![CodeQL](https://github.com/openhue/openhue-cli/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/openhue/openhue-cli/actions/workflows/github-code-scanning/codeql)
[![Maintainability with Code Climate](https://api.codeclimate.com/v1/badges/fb934bb37c36a04f8efd/maintainability)](https://codeclimate.com/github/openhue/openhue-cli/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/fb934bb37c36a04f8efd/test_coverage)](https://codeclimate.com/github/openhue/openhue-cli/test_coverage)
[![Docker Image Size (tag)](https://img.shields.io/docker/image-size/openhue/cli/latest)](https://hub.docker.com/r/openhue/cli)
[![GitHub Repo stars](https://img.shields.io/github/stars/openhue/openhue-cli)](https://github.com/openhue/openhue-cli/stargazers)

## Overview

OpenHue CLI is a command-line interface for interacting with Philips Hue smart lighting systems. 
This tool provides a convenient way to control your Philips Hue lights and perform various tasks using the command line.

[![How to setup OpenHue CLI](./docs/images/openhue_setup.gif)](https://www.openhue.io/cli/setup)

## Features

- Discover and connect to Philips Hue bridges.
- List available lights and their status.
- Control lights: on, off, brightness, and color.
- Schedule light actions.

For a complete list of features and usage, 
please refer to the [OpenHue CLI online documentation](https://www.openhue.io/cli/openhue-cli).

## Getting Started

To begin developing with Open-Hue's OpenAPI specification, follow these steps:

### Prerequisites

Before you start, ensure that you have the following prerequisites installed:
- [Golang](https://go.dev/doc/install) that is used to build and run the project
- [GoReleaser](https://goreleaser.com) (_optional_) that is used to build and release the binaries
- [Docker](https://docs.docker.com/engine/install/) (_optional_) that is used to build the CLI Docker Image and run it as a container

### Fork the Repository
Before contributing to OpenHue CLI, it's a good practice to [fork](https://github.com/openhue/openhue-cli/fork) the repository to your own GitHub account.
This will create a copy of the project that you can work on independently.

### Build

1. Clone the OpenHue CLI repository to your local machine:
```shell
git clone https://github.com/your-username/openhue-cli.git
cd openhue-cli
```
2. Run the following command to build OpenHue CLI on your local environment:
```shell
make build
```

### Test
Run the following command to execute all the tests and calculate the code coverage:
```shell
make test
```
If you want, you can also run the following command to visualize the coverage analysis in your browser: 
```shell
make coverage
```
> or use `make coverage html=true` to visualize the HTML report in your default web browser

### Generate the OpenHue API Client
Run the following command to generate the [OpenHue API Client](https://github.com/openhue/openhue-api): 
```shell
make generate
```
If there was any OpenAPI specification change, this command will update 
the [`./openhue/gen/openhue.gen.go`](./openhue/gen/openhue.gen.go) file. 
Please note that this file must never be manually edited!

You also generate the client from another spec location:
```
make generate spec=/path/to/local/openhue.yaml
```

## License
[![GitHub License](https://img.shields.io/github/license/openhue/openhue-cli)](https://github.com/openhue/openhue-cli/blob/main/LICENSE)

Open-Hue is distributed under the [Apache License 2.0](http://www.apache.org/licenses/),
making it open and free for anyone to use and contribute to.
See the [license](./LICENSE) file for detailed terms.
