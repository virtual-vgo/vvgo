![tests](https://github.com/virtual-vgo/vvgo/workflows/tests/badge.svg)
![build](https://github.com/virtual-vgo/vvgo/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/virtual-vgo/vvgo)](https://goreportcard.com/report/github.com/virtual-vgo/vvgo)

# Virtual Video Game Orchestra

:wave: We are the Virtual Video Game Orchestra (VVGO for short). Our mission is to provide a fun and accessible virtual community of musicians from around the world through performing video game music.

## Changing HTML pages

HTML pages are generated from [go templates](https://golang.org/pkg/text/template/).
These are affectionately reffered to as _views_.
Views, along with all public files are found in [here](https://github.com/virtual-vgo/vvgo/tree/master/public).

## Run VVGO locally

### 1. Install build tools

In order to build, test, and run the vvgo webapp, you will need to install git, docker, yarn, and golang.
Below are links to installation docs for each service:

#### Git
 * A version control system that we use to tracks changes to the source code.
 * Installers: [Windows](https://gitforwindows.org/) | [Mac](https://git-scm.com/download/mac) | [Linux](https://git-scm.com/download/linux)

#### Docker
 * A container engine that we use to download and run service dependencies for the webapp.
 * Installers: [Windows](https://docs.docker.com/docker-for-windows/install/) | [Mac](https://docs.docker.com/docker-for-mac/install/) | [Linux](https://docs.docker.com/engine/install/)

#### NPM
 * Manages and downloads the javascript dependencies.
 * Installers: [All](https://nodejs.org/en/download/)

#### Golang 1.14
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

### 3. Download javascript dependencies
```sh
cd public && npm install
```

### 4. Build and run the app!
```sh
go generate ./... && go build -v -o vvgo ./cmd/vvgo && ./vvgo
```

