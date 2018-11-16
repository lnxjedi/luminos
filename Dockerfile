FROM alpine:latest

ARG version=v0.9.4-snapshot

RUN wget https://github.com/lnxjedi/luminos/releases/download/${version}/luminos-linux -O /usr/local/bin/luminos \
 && chmod a+rx /usr/local/bin/luminos

USER daemon:daemon

# Best thing to do is mount a volume to /var/www, or bind mount a git repo
WORKDIR /var/www

# settings.yaml for your site should specify port: 9000
EXPOSE 9000

ENTRYPOINT /usr/local/bin/luminos -i run
