### BEGIN building
FROM golang:1-alpine AS builder

# Installing binutils for striping binaries and "build-base" for using cgo
RUN apk add --no-cache --virtual .build-essentials binutils build-base

# Setting workdir
WORKDIR /tmp/mwp_src

# Copy source files inside build container
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./api ./api
COPY ./ui/assets.go ./ui/assets.go
COPY ./ui/dist ./ui/dist

# Installing golang dependency modules
RUN go mod download
RUN go mod verify

# Building application
RUN rm -rf ./dist/
RUN RUN set -eux; \
	go build -a -x -v \
	-ldflags="-linkmode=external -s -w -extldflags=--static-pie" \
	-buildmode=pie \
	-tags=osusergo,netgo,static_build \
	-buildvcs=false \
	-o dist/mwp api/cmd/main.go
### END building

FROM alpine:3

## Github Container Registry
LABEL org.opencontainers.image.source=https://github.com/maahdima/mwp

# Installing init software
RUN set -eux; \
	apk add --no-cache --virtual .init-program catatonit && \
	ln -sf /usr/bin/catatonit /sbin/init

# ensure www-data user exists
# 82 is the standard uid/gid for "www-data" in Alpine
RUN set -eux && adduser -u 82 -D -S -s /sbin/nologin -h /var/www/mwp -G www-data www-data

# Setting workdir
WORKDIR /var/www/mwp

RUN chown -R www-data:www-data /var/www/mwp

# Copy binary files from builder container
COPY --from=builder --chown=root:root /tmp/mwp_src/dist/mwp /usr/local/sbin/

# Install binary dependencies
RUN apk add --no-cache --virtual .runtime-dependencies `scanelf --needed --nobanner --format '%n#p' --recursive /usr/local/sbin/ | tr ',' '\n' | sort -u | awk '{ print "so:" $1 }'`

# Create the peer files directory
RUN mkdir -p peer-files
RUN chown -R www-data:www-data peer-files

# Putting app version inside container just for version tracking
ARG APP_COMMIT_SHA=unknown
RUN echo "${APP_COMMIT_SHA}" > /.app_commit_sha

# Setting user and group
USER www-data:www-data

# Setting runtime environment variables
ENV MODE=production
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=3000
ENV PEER_FILES_DIR=/var/www/mwp/peer-files

# Expose port 3000 and start server
EXPOSE 3000
ENTRYPOINT ["init", "--", "/bin/sh", "-c"]
CMD ["exec mwp"]
