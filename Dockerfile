FROM --platform=$BUILDPLATFORM golang:1.24.2-alpine3.20@sha256:00f149d5963f415a8a91943531b9092fde06b596b276281039604292d8b2b9c8 AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILDFLAGS

RUN apk -U add make

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
ADD . .

# Build
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH make $BUILDFLAGS build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest

WORKDIR /

COPY --from=builder /workspace/bin/loggen .

ENTRYPOINT ["/loggen"]
