FROM alpine:3.4

RUN apk update \
  && apk --update add nodejs

RUN mkdir -p /react
WORKDIR /react

# キャッシュ効率を上げるためにまずpackage.jsonだけcopyしてnpm installする
COPY ./package.json ./npm-shrinkwrap.json /react/
RUN npm install && npm cache clean

CMD ["npm", "run", "dev"]
