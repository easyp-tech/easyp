FROM --platform=${BUILDPLATFORM} golang:1.25.2-alpine3.22 AS builder

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

FROM alpine:3.22

RUN apk update --no-cache

COPY --from=builder /easyp /easyp
