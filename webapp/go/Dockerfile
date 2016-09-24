FROM golang:alpine

RUN apk update \
  && apk --update add git

RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/jmoiron/sqlx
RUN go get goji.io/pat
RUN go get golang.org/x/net/context

RUN mkdir -p /src
WORKDIR /src

COPY . /src

RUN go build -o app .

ENTRYPOINT ["/src/app"]
