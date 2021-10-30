![tests](https://github.com/virtual-vgo/vvgo/workflows/tests/badge.svg)
![build](https://github.com/virtual-vgo/vvgo/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/virtual-vgo/vvgo)](https://goreportcard.com/report/github.com/virtual-vgo/vvgo)

# Virtual Video Game Orchestra

:wave: We are the Virtual Video Game Orchestra (VVGO for short). Our mission is to provide a fun and accessible virtual community of musicians from around the world through performing video game music.

## Run VVGO locally

### 1. Install build tools

In order to build, test, and run the vvgo webapp, you will need to install git, docker, yarn, and golang.
Below are links to installation docs for each service:

#### Git
 * A version control system that we use to tracks changes to the source code.
 * Installers: [Windows](https://gitforwindows.org/) | [Mac](https://git-scm.com/download/mac) | [Linux](https://git-scm.com/download/linux)

#### WSL 2 | Windows Only
 * This is a Linux integration layer for Windows 10 and required for Docker.
 * [Installation Docs](https://docs.microsoft.com/en-us/windows/wsl/install-win10)

#### Docker
 * A container engine that we use to download and run service dependencies for the webapp.
 * Installers: [Windows](https://docs.docker.com/docker-for-windows/install/) | [Mac](https://docs.docker.com/docker-for-mac/install/) | [Linux](https://docs.docker.com/engine/install/)

#### NPM
 * Manages and downloads the javascript dependencies.
 * Installers: [All](https://nodejs.org/en/download/)

#### Golang 1.16
 * Builds and compiles the source code.
 * Installers: [All](https://golang.org/dl/)

### 2. Clone the git repo

Clone the git repo and change to the source code directory.
Launch GitBash or your favorite terminal, and run this command:
```sh
git clone https://github.com/virtual-vgo/vvgo.git && cd vvgo
```

### 2. Launch runtime services

Redis and Minio are runtime dependencies for the webapp.
If the webapp cannot connect to Redis and Minio at startup, it will complain and exit.
These service can be started using the `docker-compose` command:
```sh
docker-compose up -d minio redis
```

### 3. Build the frontend
```sh
cd ui
npm install
npx webpack serve
```

### 4. Build the backend
```sh
go run ./tools/version
go generate ./...
go run ./cmd/vvgo
```

