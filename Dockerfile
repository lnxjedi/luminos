FROM alpine:latest

ARG version=0.9.4-snapshot

RUN wget https://github.com/lnxjedi/luminos/releases/download/${version}/luminos-linux -O /usr/local/bin/luminos \
 && chmod a+rx /usr/local/bin/luminos

USER daemon:daemon

WORKDIR /var/www

EXPOSE 9000

ENTRYPOINT /usr/local/bin/luminos -i run
