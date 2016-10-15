FROM alpine:3.4

RUN apk update \
  && apk --update add nodejs

RUN mkdir -p /app
WORKDIR /app

RUN npm install -g nodemon

# キャッシュ効率を上げるためにまずpackage.jsonだけcopyしてnpm installする
COPY ./package.json /app/
RUN npm install && npm cache clean

CMD ["nodemon", "--watch", "/app/", "--exec", "/app/bin/run"]
