FROM --platform=${BUILDPLATFORM} golang:1.25-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -trimpath -o /easyp ./cmd/easyp

FROM golang:1.25-alpine

RUN apk add --no-cache ca-certificates tzdata git bash

COPY --from=builder /easyp /easyp

ENTRYPOINT ["/easyp"]
