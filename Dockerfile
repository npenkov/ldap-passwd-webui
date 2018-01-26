FROM alpine:3.7

WORKDIR /app

ADD ldap-pass-webui /app/ldap-pass-webui
ADD static /app/static
ADD templates /app/templates
RUN chmod +x /app/ldap-pass-webui

EXPOSE 8080

ENTRYPOINT [ "/app/ldap-pass-webui" ]