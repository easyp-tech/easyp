FROM golang:1.22.3-alpine3.19 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

ENV GOOS linux

RUN apk update --no-cache

WORKDIR /build

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /easyp  ./cmd/easyp

FROM alpine:3.20.0

RUN apk update --no-cache && apk add --no-cache ca-certificates=20240226-r0

COPY --from=builder /easyp /easyp
