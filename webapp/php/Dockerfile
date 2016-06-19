FROM php:alpine

RUN mkdir /myapp
WORKDIR /myapp
ADD . /myapp

ENTRYPOINT php -S 0.0.0.0:9901 index.php

EXPOSE 9901
