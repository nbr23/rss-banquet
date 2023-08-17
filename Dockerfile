FROM golang:alpine as builder

WORKDIR /build

RUN apk add gcc musl-dev
COPY go* main.go modules.go config.go .
COPY parser parser

RUN go build -trimpath -o /build/atomic-banquet

FROM alpine

COPY --from=builder /build/atomic-banquet /usr/bin/atomic-banquet

CMD atomic-banquet