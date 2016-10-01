# ISUCON6 final benchmarker

```
make
```

## Bench

```
./local-bench -host 127.0.0.1 -audience1 http://127.0.0.1:8080/ -timeout 30
```

## audience

特定のRoomに入室するクライアントをいっぱい生成するウェブAPI。

```
./local-audience -initialWatcherNum 5 -watcherIncreaseInterval 5 -timeout 30 -listen 0.0.0.0:10080
```

curlで試しに実行する

```
curl 'http://127.0.0.1:10080/?baseURL=https%3A%2F%2F127.0.0.1&room=100'
```
