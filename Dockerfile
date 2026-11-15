FROM golang:1.27.1-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /reverse-proxy ./cmd/server

FROM alpine:3.22

COPY --from=build /reverse-proxy /reverse-proxy

ENV PORT=8081
EXPOSE 8081

ENTRYPOINT ["/reverse-proxy"]
