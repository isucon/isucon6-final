FROM golang:1.7.3-alpine

RUN apk update \
  && apk --update add git

RUN go get github.com/golang/dep/...

# キャッシュ効率を上げるためにglideだけ先にcopyしてインストールする
COPY ./lock.json /go/src/github.com/isucon/isucon6-final/webapp/go/
COPY ./manifest.json /go/src/github.com/isucon/isucon6-final/webapp/go/
WORKDIR /go/src/github.com/isucon/isucon6-final/webapp/go

RUN dep ensure

COPY ./ /go/src/github.com/isucon/isucon6-final/webapp/go/

RUN go build -o app .

CMD ["/go/src/github.com/isucon/isucon6-final/webapp/go/app"]
