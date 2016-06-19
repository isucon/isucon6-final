FROM alpine:3.4

RUN apk update \
  && apk --update add nodejs

RUN mkdir /myapp
WORKDIR /myapp
ADD package.json /myapp/package.json
RUN npm install --production
ADD . /myapp

ENTRYPOINT $(npm bin)/babel-node server.jsx

EXPOSE 8800
