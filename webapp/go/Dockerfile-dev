FROM golang:1.7.3-alpine

RUN apk update \
  && apk --update add git

RUN go get github.com/Masterminds/glide && \
  go get github.com/cortesi/modd/cmd/modd

# 開発環境ではこのディレクトリをマウントする
WORKDIR /go/src/github.com/isucon/isucon6-final/webapp/go

CMD ["modd"]
