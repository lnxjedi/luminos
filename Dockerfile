FROM alpine:latest

ADD https://github.com/lnxjedi/luminos/releases/download/v0.9.1/luminos-linux /usr/local/bin/luminos

RUN chmod a+rx /usr/local/bin/luminos

WORKDIR /var/www

EXPOSE 9000

#ENTRYPOINT [ "/usr/local/bin/luminos" "run" "-i" ]
