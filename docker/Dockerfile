FROM alpine:3

## Github Container Registry
LABEL org.opencontainers.image.source=https://github.com/Maahdima/MWPanel

# Installing init software
RUN set -eux; \
	apk add --no-cache --virtual .init-program catatonit && \
	ln -sf /usr/bin/catatonit /sbin/init

# ensure www-data user exists
# 82 is the standard uid/gid for "www-data" in Alpine
RUN set -eux && \
	adduser -u 82 -D -S -s /sbin/nologin -h /var/www/mwp -G www-data www-data && \
	mkdir -p /var/www/mwp && \
	chown www-data:www-data /var/www/mwp

# Setting workdir
WORKDIR /var/www/mwp
VOLUME /var/www/mwp

# Copy binary file
ARG TARGETOS
ARG TARGETARCH
COPY --chown=root:root ./build/mwp.$TARGETOS.$TARGETARCH /usr/local/sbin/mwp

# Install binary dependencies
RUN set -eux; \
    RUNTIME_DEPS=$( \
      scanelf --needed --nobanner --format '%n#p' --recursive /usr/local/sbin/ | \
      tr ',' '\n' | \
      sort -u | \
      awk 'system("[ -e /usr/local/lib/" $1 " ]") == 0 { next } { print "so:" $1 }'\
    ) && \
    apk add --no-cache --virtual .runtime-dependencies $(echo $RUNTIME_DEPS | xargs)

# Create the peer files directory
RUN set -eux; \
	mkdir -p peer-files && \
    chown www-data:www-data peer-files

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
