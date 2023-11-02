# OpenHue CLI
[![Build](https://github.com/openhue/openhue-cli/actions/workflows/build.yml/badge.svg)](https://github.com/openhue/openhue-cli/actions/workflows/build.yml)
[![CodeQL](https://github.com/openhue/openhue-cli/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/openhue/openhue-cli/actions/workflows/github-code-scanning/codeql)
[![Maintainability with Code Climate](https://api.codeclimate.com/v1/badges/fb934bb37c36a04f8efd/maintainability)](https://codeclimate.com/github/openhue/openhue-cli/maintainability)[![Docker Image Size (tag)](https://img.shields.io/docker/image-size/openhue/cli/latest)](https://hub.docker.com/r/openhue/cli)
[![GitHub Repo stars](https://img.shields.io/github/stars/openhue/openhue-cli)](https://github.com/openhue/openhue-cli/stargazers)
[![GitHub issues](https://img.shields.io/github/issues/openhue/openhue-cli)](https://github.com/openhue/openhue-cli/issues)
[![GitHub](https://img.shields.io/github/license/openhue/openhue-cli)](https://github.com/openhue/openhue-cli/blob/main/LICENSE)

## Overview

OpenHue CLI is a command-line interface for interacting with Philips Hue smart lighting systems. 
This tool provides a convenient way to control your Philips Hue lights and perform various tasks using the command line.

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

### Generate the OpenHue API Client
Run the following command to generate the [OpenHue API Client](https://github.com/openhue/openhue-api): 
```shell
make generate
```
If there was any OpenAPI specification change, this command will update 
the [`./openhue/openhue.gen.go`](./openhue/openhue.gen.go) file. 
Please note that this file must never be manually edited!

## License

Open-Hue is distributed under the [Apache License 2.0](http://www.apache.org/licenses/),
making it open and free for anyone to use and contribute to.
See the [license](./LICENSE) file for detailed terms.
