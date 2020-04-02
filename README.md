![tests](https://github.com/virtual-vgo/vvgo/workflows/tests/badge.svg)
![build](https://github.com/virtual-vgo/vvgo/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/virtual-vgo/vvgo)](https://goreportcard.com/report/github.com/virtual-vgo/vvgo)

# Virtual Video Game Orchestra

:wave: We are the Virtual Video Game Orchestra (VVGO for short), an orchestra organized by members from various IRL VGOs/GSOs, and comprised of local musicians hailing from across the globe!

## Build

You can build the webserver either using docker or go build tools. 
When you run it, you can visit the site at http://localhost:8080.

### Build with docker

```sh
# Clone the repo
git clone https://github.com/virtual-vgo/vvgo.git
# Build the docker image
docker build --tag vvgo .
# Start the container
docker run -p8080:8080 --rm vvgo
```

### Build with go

```sh
# Clone the repo
git clone https://github.com/virtual-vgo/vvgo.git && cd vvgo
# Build it
go build -o vvgo
# Run it
./vvgo
```
