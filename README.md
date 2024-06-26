<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/w/goheft/ci"><img src="https://kaos.sh/w/goheft/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/r/goheft"><img src="https://kaos.sh/r/goheft.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/b/goheft"><img src="https://kaos.sh/b/43c7247d-ff5d-4684-8d9d-cf5e85b8c7a7.svg" alt="codebeat badge" /></a>
  <a href="https://kaos.sh/w/goheft/codeql"><img src="https://kaos.sh/w/goheft/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#usage-demo">Usage demo</a> • <a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`goheft` is simple utility for listing sizes of all used static libraries compiled into golang binary.

### Usage demo

[![demo](https://gh.kaos.st/goheft-070.gif)](#usage-demo)

### Installation

#### From source

To build the GoHeft from scratch, make sure you have a working Go 1.19+ workspace ([instructions](https://go.dev/doc/install)), then:

```
go install github.com/essentialkaos/goheft@latest
```

#### Prebuilt binaries

You can download prebuilt binaries for Linux and macOS from [EK Apps Repository](https://apps.kaos.st/goheft/):

```bash
bash <(curl -fsSL https://apps.kaos.st/get) goheft
```

### Command-line completion

You can generate completion for `bash`, `zsh` or `fish` shell.

Bash:
```bash
sudo goheft --completion=bash 1> /etc/bash_completion.d/goheft
```


ZSH:
```bash
sudo goheft --completion=zsh 1> /usr/share/zsh/site-functions/goheft
```


Fish:
```bash
sudo goheft --completion=fish 1> /usr/share/fish/vendor_completions.d/goheft.fish
```

### Man documentation

You can generate man page for goheft using next command:

```bash
goheft --generate-man | sudo gzip > /usr/share/man/man1/goheft.1.gz
```

### Usage

<p align="center"><img src=".github/images/usage.svg"/></p>

### Build Status

| Branch | Status |
|--------|--------|
| `master` | [![CI](https://kaos.sh/w/goheft/ci.svg?branch=master)](https://kaos.sh/w/goheft/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/goheft/ci.svg?branch=develop)](https://kaos.sh/w/goheft/ci?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
