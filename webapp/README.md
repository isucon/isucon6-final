# 起動方法

例えばPHP実装の場合は

```sh
ln -s docker-compose-php.yml docker-compose.yml
docker-compose up --build
```

でポート443で起動し、 https://localhost/ にアクセスできるようになります。

`-dev` が付いているものは開発用です。

```sh
ln -s docker-compose-php-dev.yml docker-compose.yml
docker-compose up --build
```

開発用はdockerのホストマシン側のソースコードのディレクトリをマウントしており、ホストマシンでコードを変更すると自動的にアプリケーションがリロードされるようになっています。
