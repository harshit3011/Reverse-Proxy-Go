# Reverse Proxy and Load Balancer

A small Go HTTP reverse proxy that sends incoming requests to configured origin
servers in round-robin order.

## Requirements

- Go 1.27.1 or later
- Docker, if running the container

## Configuration

The server reads configuration from environment variables. At startup it also
loads a `.env` file from the current working directory if one exists. `.env` is
ignored by Git; do not commit secrets.

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `SECRET_TOKEN` | Yes | — | Expected value of the request's `Authorization` header |
| `PORT` | No | `8081` | Port the proxy listens on |
| `ORIGIN_SERVERS` | No | `http://127.0.0.1:8000` through `http://127.0.0.1:8004` | Comma-separated origin URLs |

Example `.env`:

```env
SECRET_TOKEN=mysecrettoken
PORT=8081
ORIGIN_SERVERS=http://127.0.0.1:8000,http://127.0.0.1:8001
```

If `ORIGIN_SERVERS` is omitted, the program uses ports 8000 through 8004 on
localhost. Make sure the configured origin servers are running and reachable.

## Run locally

Start one or more HTTP origin servers in separate terminals. For example, from
the repository root, this starts a simple server on port 8000:

```powershell
python -m http.server 8000
```

Set the proxy configuration and run the server from the repository root:

```powershell
$env:SECRET_TOKEN="mysecrettoken"
$env:PORT="8081"
$env:ORIGIN_SERVERS="http://127.0.0.1:8000"
go run .\cmd\server
```

Send a request from another terminal:

```powershell
curl.exe -i -H "Authorization: mysecrettoken" "http://localhost:8081/"
```

The proxy forwards the request path, query string, method, headers, and body to
the selected origin, then relays the origin's response. An incorrect or missing
token returns `401 Unauthorized`. If the origin cannot be reached, the proxy
returns `502 Bad Gateway`.

## Run with Docker

Build the image from the repository root:

```powershell
docker build -t reverse-proxy-go .
```

If an origin server is running on the Windows host, start the container with:

```powershell
docker run --rm -p 8081:8081 `
  -e SECRET_TOKEN=mysecrettoken `
  -e ORIGIN_SERVERS=http://host.docker.internal:8000 `
  reverse-proxy-go
```

Then send the same `curl.exe` request shown above. `host.docker.internal`
allows the container to reach a service running on the host. For origins in
other containers, put the containers on the same Docker network and use the
origin container's network name and port.

## Checks

Run the Go package tests from the repository root:

```powershell
go test ./...
```
