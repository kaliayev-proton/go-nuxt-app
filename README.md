# Simple Go API

## Get started

We init a Go project:
`go mod init ambassador`

### Download Fiber package with Go:

`go get github.com/gofiber/fiber/v2`

And it is specified in the `go.mod` file

## Setup with Docker

`sudo make up` (Linux)

## Live reload for Go and Docker

[Live reload](https://github.com/cosmtrek/air)

We add to the Docker file:
Presinstall: `curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin`

## Is you want to rebuild the Docker image:

`docker-compose up --build`

## Execute commands into Docker

1. Connect via sh with **backend** container with docker-compose: `docker-compose exec backend sh`
