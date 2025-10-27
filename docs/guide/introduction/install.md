# Install the EasyP cli

### Install from system's package manager

##### Recommended installation method

#### Homebrew

You can install a binary release on macOS using brew:

```bash
brew install easyp-tech/tap/easyp
```

`Work in progress`
apt, yum, pacman, etc.

## Local Installation

#### _**Not recommended**_

Local installation is not recommended because package managers provide:
- Automatic updates and security patches
- Better integration with system package management
- Easier installation and removal process
- Verified and signed binaries

#### Requires Go 1.24 or later

### Go install

It will install the latest stable version of easyp.

```bash
go install github.com/easyp-tech/easyp/cmd/easyp@latest
```

### Build from source


1. Clone repository

```bash
git clone https://github.com/easyp-tech/easyp.git
```

2. Build by golang

```bash
go build ./cmd/easyp
```

