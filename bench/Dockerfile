FROM golang:alpine

RUN \
  apk add --no-cache curl git make && \
  curl -sL https://glide.sh/get | sh

WORKDIR ${GOPATH}/src/github.com/catatsuy/isucon6-final/bench
COPY . ${GOPATH}/src/github.com/catatsuy/isucon6-final/bench

RUN \
  glide update && \
  go get -u github.com/jteeuwen/go-bindata/... && \
  make

#CMD ["./local-bench", "-urls", "https://react", "-timeout", "30"]
