FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

ENV GOOS linux

RUN apk update --no-cache

WORKDIR /build

ADD go.mod .

ADD go.sum .

RUN go mod download

COPY . .

RUN go build -o /easyp  ./cmd/easyp

FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates

COPY --from=builder /easyp /easyp
