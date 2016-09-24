FROM golang:alpine

RUN apk update \
  && apk --update add git

RUN go get github.com/Masterminds/glide

RUN mkdir -p /go/src/github.com/catatsuy/isucon6-final/webapp/go
WORKDIR /go/src/github.com/catatsuy/isucon6-final/webapp/go

COPY ./ /go/src/github.com/catatsuy/isucon6-final/webapp/go/
RUN glide install

RUN go build -o app .

ENTRYPOINT ["/go/src/github.com/catatsuy/isucon6-final/webapp/go/app"]
