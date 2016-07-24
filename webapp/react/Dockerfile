FROM alpine:3.4

RUN apk update \
  && apk --update add nodejs

VOLUME ["/app"]

ENTRYPOINT ["node", "/app/server.js"]
