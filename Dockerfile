FROM --platform=${BUILDPLATFORM} golang:1.26-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG RELEASE_VERSION

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -trimpath -ldflags "-s -w -X github.com/easyp-tech/easyp/internal/version.releaseVersion=${RELEASE_VERSION}" -o /easyp ./cmd/easyp

FROM alpine:3.22

# Keep package pins aligned with the Alpine release; override for verified updates.
ARG CA_CERTIFICATES_VERSION=20260909-r0
ARG TZDATA_VERSION=2026d-r0
ARG GIT_VERSION=2.49.1-r0
ARG BASH_VERSION=5.2.37-r0

RUN apk add --no-cache \
    ca-certificates="${CA_CERTIFICATES_VERSION}" \
    tzdata="${TZDATA_VERSION}" \
    git="${GIT_VERSION}" \
    bash="${BASH_VERSION}"

COPY --from=builder /easyp /easyp

ENTRYPOINT ["/easyp"]
