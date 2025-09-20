FROM --platform=${BUILDOS}/${BUILDARCH} golang:alpine AS builder

WORKDIR /build

RUN apk add gcc musl-dev
COPY go* *.go /build/
COPY static static
COPY parser parser
COPY style style
COPY utils utils
COPY config config

RUN GOOS=linux GOARCH=arm64 go build -trimpath -o rss-banquet-linux-arm64
RUN GOOS=linux GOARCH=amd64 go build -trimpath -o rss-banquet-linux-amd64

FROM builder AS test

COPY testsuite testsuite
RUN apk add --no-cache ca-certificates && rm -rf /var/cache/apk/*

# Base

FROM --platform=${TARGETOS}/${TARGETARCH} alpine:latest AS base
ARG TARGETARCH
ARG TARGETOS

RUN apk add --no-cache ca-certificates && rm -rf /var/cache/apk/*

COPY --from=builder /build/rss-banquet-${TARGETOS}-${TARGETARCH} /usr/bin/rss-banquet

# Server

FROM base AS server
ENV PORT=8080
ENV GIN_MODE=release

EXPOSE ${PORT}

CMD rss-banquet server -p ${PORT}

# Development

FROM builder AS dev-server
ENV PORT=8080
ENV GIN_MODE=debug
EXPOSE ${PORT}

RUN apk update && apk add watchexec

CMD ["watchexec", "-w", ".", "-e", "go,sum,mod", "-r", "sh", "-c", "date && echo [WATCHEXEC] Building... && go build && echo [WATCHEXEC] Built, launching && ./rss-banquet server -p ${PORT}"]

# nginx

FROM base AS nginx
ENV PORT=8080
ENV GIN_MODE=release

EXPOSE ${PORT}

RUN apk update && apk add nginx && rm -rf /var/cache/apk/*

RUN cat <<EOF > /etc/nginx/http.d/default.conf
proxy_cache_path /var/lib/nginx/cache levels=1:2 keys_zone=mycache:50m max_size=1g inactive=15m use_temp_path=off;

gzip on;
gzip_types application/json application/rss+xml application/atom+xml;

  server {
		listen ${PORT};

		proxy_cache mycache;

		location / {
			proxy_pass http://localhost:8081;
			proxy_cache_valid 200 15m;
			proxy_cache_lock on;
		}
	}

EOF

CMD ["sh", "-c", "nginx && rss-banquet server -p 8081"]
