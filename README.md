<p align="center"><a href="#readme"><img src="https://gh.kaos.st/goheft.svg"/></a></p>

<p align="center">
  <a href="https://github.com/essentialkaos/goheft/actions"><img src="https://github.com/essentialkaos/goheft/workflows/CI/badge.svg" alt="GitHub Actions Status" /></a>
  <a href="https://github.com/essentialkaos/goheft/actions?query=workflow%3ACodeQL"><img src="https://github.com/essentialkaos/goheft/workflows/CodeQL/badge.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/essentialkaos/goheft"><img src="https://goreportcard.com/badge/github.com/essentialkaos/goheft"></a>
  <a href="https://codebeat.co/projects/github-com-essentialkaos-goheft-master"><img alt="codebeat badge" src="https://codebeat.co/badges/43c7247d-ff5d-4684-8d9d-cf5e85b8c7a7" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`goheft` is simple utility for listing sizes of all used static libraries compiled into golang binary.

### Screenshots

![Screenshot](https://gh.kaos.st/goheft.png)

### Installation

#### From source

To build the GoHeft from scratch, make sure you have a working Go 1.13+ workspace ([instructions](https://golang.org/doc/install)), then:

```
go get github.com/essentialkaos/goheft
```

If you want update GoHeft to latest stable release, do:

```
go get -u github.com/essentialkaos/goheft
```

#### Prebuilt binaries

You can download prebuilt binaries for Linux and OS X from [EK Apps Repository](https://apps.kaos.st/goheft/):

```bash
bash <(curl -fsSL https://apps.kaos.st/get) goheft
```

### Usage

```
Usage: goheft {options} file

Options

  --external, -e         Shadow internal packages
  --min-size, -m size    Don't show with size less than defined
  --raw, -r              Print raw data
  --no-color, -nc        Disable colors in output
  --help, -h             Show this help message
  --version, -v          Show version

Examples

  goheft application.go
  Show size of each used library

  goheft application.go -m 750kb
  Show size of each used library which greater than 750kb

```

### Build Status

| Branch | Status |
|------------|--------|
| `master` | [![CI](https://github.com/essentialkaos/goheft/workflows/CI/badge.svg?branch=master)](https://github.com/essentialkaos/goheft/actions) |
| `develop` | [![CI](https://github.com/essentialkaos/goheft/workflows/CI/badge.svg?branch=develop)](https://github.com/essentialkaos/goheft/actions) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
