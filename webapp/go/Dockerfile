FROM golang:1.7.3-alpine

RUN apk update \
  && apk --update add git

RUN go get github.com/Masterminds/glide

# キャッシュ効率を上げるためにglideだけ先にcopyしてインストールする
COPY ./glide.yaml /go/src/github.com/isucon/isucon6-final/webapp/go/
COPY ./glide.lock /go/src/github.com/isucon/isucon6-final/webapp/go/
WORKDIR /go/src/github.com/isucon/isucon6-final/webapp/go

RUN glide install

COPY ./ /go/src/github.com/isucon/isucon6-final/webapp/go/

RUN go build -o app .

CMD ["/go/src/github.com/isucon/isucon6-final/webapp/go/app"]
