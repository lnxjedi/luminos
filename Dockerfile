FROM alpine:latest

ARG version=v0.9.4-snapshot

ADD https://github.com/lnxjedi/luminos/releases/download/${version}/luminos-linux /usr/local/bin/luminos

RUN chmod a+rx /usr/local/bin/luminos

USER daemon:daemon

# This is where the site should go
WORKDIR /var/www

ENTRYPOINT [ "/usr/local/bin/luminos" ]
