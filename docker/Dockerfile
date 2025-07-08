### BEGIN building
FROM golang:1-alpine AS builder

# Installing binutils for striping atlas binary and "build-base" for using cgo
RUN apk add --no-cache --virtual .build-essentials binutils build-base

# Copy source files inside build container
COPY ./api /tmp/mw-api_src
WORKDIR /tmp/mw-api_src

# Installing golang dependency modules
RUN go mod download
RUN go mod verify

# Building application
RUN rm -rf ./dist/
RUN go build -ldflags="-linkmode=external -s -w -extldflags=--static-pie" -buildmode=pie -tags=osusergo,netgo,static_build -buildvcs=false -a -x -v -o dist/mw-api cmd/main.go
RUN set -eux; \
	CPU_ARCH=`(test $(uname -m) = 'aarch64' && echo 'arm64') || (test $(uname -m) = 'x86_64' && echo 'amd64') || (echo 'Unsupported architecture' && exit 1)` && \
	wget -qO ./dist/atlas https://release.ariga.io/atlas/atlas-community-linux-${CPU_ARCH}-latest && \
	echo "$(wget -qO- https://release.ariga.io/atlas/atlas-community-linux-${CPU_ARCH}-latest.sha256)  ./dist/atlas" | sha256sum -swc && \
	strip -vps ./dist/atlas && \
	chmod +x ./dist/atlas
### END building

FROM alpine:3

# Installing init software
RUN apk add --no-cache --virtual .init-program catatonit
RUN ln -sf /usr/bin/catatonit /sbin/init

# ensure www-data user exists
# 82 is the standard uid/gid for "www-data" in Alpine
RUN set -eux && adduser -u 82 -D -S -s /sbin/nologin -h /var/www -G www-data www-data

RUN chown -R www-data:www-data /var/www

# Setting workdir
WORKDIR /var/www

# Copy binary files from builder container
RUN mkdir -p /usr/local/share/mw-api/migrations
COPY --from=builder --chown=root:root /tmp/mw-api_src/dataservice/migration[s]/ /usr/local/share/mw-api/migrations/
COPY --from=builder --chown=root:root /tmp/mw-api_src/dist/mw-api /usr/local/sbin/
COPY --from=builder --chown=root:root /tmp/mw-api_src/dist/atlas /usr/local/sbin/

# Install binary dependencies
RUN apk add --no-cache --virtual .runtime-dependencies `scanelf --needed --nobanner --format '%n#p' --recursive /usr/local/bin | tr ',' '\n' | sort -u | awk '{ print "so:" $1 }'`

# Copying the ui static files from the build directory to the public directory
COPY --chown=root:root ui/dist public

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
ENV MIGRATION_DIR=/usr/local/share/mw-api/migrations
ENV PUBLIC_DIR=/var/www/public
ENV PEER_FILES_DIR=/var/www/peer-files

# Expose port 3000 and start server
EXPOSE 3000
ENTRYPOINT ["init", "--", "/bin/sh", "-c"]
CMD ["exec mw-api"]
