# ISUCON6 React Frontend

## development

Use https://github.com/cortesi/modd to build everyting into `build/`

```
$ go get github.com/cortesi/modd/cmd/modd
$ npm install
$ modd
```

`modd` starts a build server.

To start a development server which mounts the build directory, do the following.

```
$ cd ..
$ ln -snf docker-compose-php-dev.yml docker-compose.yml # for example
$ docker-compose up --build
```

## production

Instead, to start a production server which copies the source and build inside it, do the following.

```
$ cd ..
$ ln -snf docker-compose-php.yml docker-compose.yml # for example
$ docker-compose up --build
```
