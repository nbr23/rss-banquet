FROM --platform=${BUILDOS}/${BUILDARCH} golang:alpine as builder

WORKDIR /build

RUN apk add gcc musl-dev
COPY go* main.go modules.go config.go .
COPY parser parser

RUN GOOS=linux GOARCH=arm64 go build -trimpath -o atomic-banquet-linux-arm64
RUN GOOS=linux GOARCH=amd64 go build -trimpath -o atomic-banquet-linux-amd64

FROM --platform=${TARGETOS}/${TARGETARCH} alpine:latest
ARG TARGETARCH
ARG TARGETOS

COPY --from=builder /build/atomic-banquet-${TARGETOS}-${TARGETARCH} /usr/bin/atomic-banquet

CMD atomic-banquet