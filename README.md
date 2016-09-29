# ISUCON6 final

```
make
```

## audience

特定のRoomに入室するクライアントをいっぱい生成するウェブAPI。

```
./local-audience -initialWatcherNum 5 -watcherIncreaseInterval 5 -timeout 60 -listen 0.0.0.0:10080 -targetScheme https
```

```
curl 'http://127.0.0.1:10080/?target=127.0.0.1&room=100'
```
