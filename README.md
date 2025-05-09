<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/w/goheft/ci"><img src="https://kaos.sh/w/goheft/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/r/goheft"><img src="https://kaos.sh/r/goheft.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/w/goheft/codeql"><img src="https://kaos.sh/w/goheft/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#usage-demo">Usage demo</a> • <a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`goheft` is simple utility for listing sizes of all used static libraries compiled into golang binary.

### Usage demo

[![demo](https://github.com/user-attachments/assets/66d24052-6cc7-4537-8871-82180ca0e4f9)](#usage-demo)

### Installation

#### From source

To build the GoHeft from scratch, make sure you have a working [Go 1.23+](https://github.com/essentialkaos/.github/blob/master/GO-VERSION-SUPPORT.md) workspace ([instructions](https://go.dev/doc/install)), then:

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

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/.github/blob/master/CONTRIBUTING.md).

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://kaos.dev"><img src="https://raw.githubusercontent.com/essentialkaos/.github/refs/heads/master/images/ekgh.svg"/></a></p>
