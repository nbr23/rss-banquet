FROM --platform=${BUILDOS}/${BUILDARCH} golang:alpine as builder

WORKDIR /build

RUN apk add gcc musl-dev
COPY go* main.go modules.go config.go /build/
COPY parser parser

RUN GOOS=linux GOARCH=arm64 go build -trimpath -o atomic-banquet-linux-arm64
RUN GOOS=linux GOARCH=amd64 go build -trimpath -o atomic-banquet-linux-amd64

# Fetcher

FROM --platform=${TARGETOS}/${TARGETARCH} alpine:latest as fetcher
ARG TARGETARCH
ARG TARGETOS

COPY --from=builder /build/atomic-banquet-${TARGETOS}-${TARGETARCH} /usr/bin/atomic-banquet

CMD atomic-banquet fetcher

# Server

FROM fetcher as server
ENV PORT 8080
ENV GIN_MODE release

EXPOSE ${PORT}

CMD atomic-banquet server -p ${PORT}

# Development

FROM builder as dev-server
ENV PORT 8080
ENV GIN_MODE debug
EXPOSE ${PORT}

RUN apk update && apk add watchexec

CMD watchexec -w . -e go,sum,mod -r sh -c "date && echo [WATCHEXEC] Building... && go build && echo [WATCHEXEC] Built, launching && ./atomic-banquet server -p ${PORT}"

# nginx

FROM fetcher as nginx
ENV PORT 8080
ENV GIN_MODE release

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

CMD ["sh", "-c", "nginx && atomic-banquet server -p 8081"]
