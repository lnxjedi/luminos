FROM alpine:latest

COPY luminos-linux /usr/local/bin/luminos

RUN chmod +x /usr/local/bin/luminos

WORKDIR /var/www

EXPOSE 8080

ENTRYPOINT [ "/usr/local/bin/luminos" ]
