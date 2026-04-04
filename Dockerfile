FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/easydocker ./cmd/easydocker

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/easydocker /usr/local/bin/easydocker

ENTRYPOINT ["/usr/local/bin/easydocker"]