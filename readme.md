# GoHeft [![Build Status](https://travis-ci.org/essentialkaos/goheft.svg?branch=master)](https://travis-ci.org/essentialkaos/goheft) [![Go Report Card](https://goreportcard.com/badge/github.com/essentialkaos/goheft)](https://goreportcard.com/report/github.com/essentialkaos/goheft) [![License](https://gh.kaos.io/ekol.svg)](https://essentialkaos.com/ekol)

`goheft` is simple utility for listing sizes of all used static libraries compiled into golang binary.

![Screenshot](https://gh.kaos.io/goheft.png)

## Installation

Before the initial install allows git to use redirects for [pkg.re](https://github.com/essentialkaos/pkgre) service (reason why you should do this described [here](https://github.com/essentialkaos/pkgre#git-support)):

```
git config --global http.https://pkg.re.followRedirects true
```

To build the GoHeft from scratch, make sure you have a working Go 1.7+ workspace ([instructions](https://golang.org/doc/install)), then:

```
go get github.com/essentialkaos/goheft
```

If you want update GoHeft to latest stable release, do:

```
go get -u github.com/essentialkaos/goheft
```

## Usage

```
Usage: goheft {options} file

Options

  --external, -e         Shadow internal packages
  --min-size, -m size    Don't show with size less than defined
  --no-color, -nc        Disable colors in output
  --help, -h             Show this help message
  --version, -v          Show version

Examples

  goheft application.go
  Show size of each used library

  goheft application.go -m 750kb
  Show size of each used library which greater than 750kb

```

## Build Status

| Branch | Status |
|------------|--------|
| `master` | [![Build Status](https://travis-ci.org/essentialkaos/goheft.svg?branch=master)](https://travis-ci.org/essentialkaos/goheft) |
| `develop` | [![Build Status](https://travis-ci.org/essentialkaos/goheft.svg?branch=develop)](https://travis-ci.org/essentialkaos/goheft) |

## Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

## License

[EKOL](https://essentialkaos.com/ekol)
