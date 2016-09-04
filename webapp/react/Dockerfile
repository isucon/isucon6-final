FROM alpine:3.4

RUN apk update \
  && apk --update add nodejs

RUN mkdir -p /react
WORKDIR /react

# キャッシュ効率を上げるためにまずpackage.jsonだけcopyしてnpm installする
COPY ./package.json /react/package.json
RUN npm install

# その後他のファイルも含めてcopyしてbabelやwebpackの生成を行う
COPY . /react
RUN npm run build

ENTRYPOINT ["node", "/react/build/server.js"]
