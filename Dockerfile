FROM --platform=$BUILDPLATFORM golang:1.20-alpine3.18@sha256:7efb78dac256c450d194e556e96f80936528335033a26d703ec8146cec8c2090 AS builder

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
