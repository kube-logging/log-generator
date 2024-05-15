FROM --platform=$BUILDPLATFORM golang:1.22-alpine3.18@sha256:d1a601b64de09e2fa38c95e55838961811d5ca11062a8f4230a5c434b3ae2a34 as builder

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
