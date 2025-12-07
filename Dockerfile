FROM --platform=${BUILDPLATFORM} golang:1.25-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0

ENV GOOS=linux

RUN apk update --no-cache

WORKDIR /build

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /easyp  ./cmd/easyp

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /easyp /easyp

ENTRYPOINT ["/easyp"]
