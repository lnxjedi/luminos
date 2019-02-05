FROM alpine:latest

ARG version=v0.9.4-snapshot

ADD https://github.com/lnxjedi/luminos/releases/download/${version}/luminos-linux /usr/local/bin/luminos

RUN chmod a+rx /usr/local/bin/luminos

USER daemon:daemon

# This is where the site should go
WORKDIR /var/www

# settings.yaml for your site should specify port: 9000
EXPOSE 9000

ENTRYPOINT /usr/local/bin/luminos -i run
