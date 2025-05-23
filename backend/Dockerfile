ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o bin/server cmd/server/server.go

FROM debian:bookworm

RUN apt-get update && apt-get install -y curl
RUN curl -fsSL https://github.com/pressly/goose/releases/download/v3.24.3/goose_linux_x86_64 -o /usr/local/bin/goose && \
    chmod +x /usr/local/bin/goose
RUN apt-get update && apt-get install -y make
COPY --from=builder /usr/src/app/bin/server /usr/local/bin/
COPY --from=builder /usr/src/app/Makefile /usr/local/bin/
COPY --from=builder /usr/src/app/migrations /usr/local/bin/migrations

WORKDIR /usr/local/bin
CMD ["/bin/bash", "-c", "make migrate && server"]
