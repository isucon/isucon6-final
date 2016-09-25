FROM php:apache

RUN docker-php-ext-install pdo pdo_mysql

RUN apt-get update && apt-get install -y unzip

RUN a2enmod rewrite

# キャッシュ効率を上げるためにcomposerだけ先にcopyしてインストールする
COPY ./composer.phar /var/www
COPY ./composer.json /var/www
COPY ./composer.lock /var/www
WORKDIR /var/www

RUN ./composer.phar install --no-dev

COPY . /var/www
