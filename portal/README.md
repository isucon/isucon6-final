# isucon6f-portal

ISUCON6 本選ポータルサイトです。

## 運営アカウント

どの日も同じアカウントで入れます。

- ID: 9999
- PASS: `Btw5R5fskVvXOzT`

他のチームと同様の扱いでログインできます。

## デプロイ

~~~
Host isucon6f-portal
    User isucon
    HostName 13.78.94.217
~~~

で

    make deploy TARGET=isucon6f-portal ANSIBLE_ARGS=-vv

## 起動オプション

- `-database-dsn <dsn="root:@/isu6fportal">`
- `-starts-at <hour=10>`
- `-ends-at <hour=18>`

## 運用

本選終了後はジョブのエンキューやログインができなくなりますが、スコア等は見えます。

終了後5分ぐらいたったらnginxでBASIC認証をかけ、`-ends-at=-1` で再起動し、各チームでログインしてベンチマークを実行することで追試できます。

## 開発・運用むけ情報

秘密のURLです。

```
user: isucon6f
pass: X1grZy5vTrIRONRmsTKl
```

- /mBGWHqBVEjUSKpBF/debug/queue キュー一覧（標準エラー確認など）
- /mBGWHqBVEjUSKpBF/messages トップページ等に表示するメッセージ管理画面
- /mBGWHqBVEjUSKpBF/proxy/nginx.conf proxyが持つnginx.conf
- /mBGWHqBVEjUSKpBF/debug/vars Goのデバッグ情報
- /mBGWHqBVEjUSKpBF/debug/leaderboard 17時以降も更新される管理用リーダーボード
- /mBGWHqBVEjUSKpBF/debug/proxies 登録されているproxy一覧

## ローカルで開発する

```
mysql -uroot -e 'DROP DATABASE IF EXISTS isu6fportal;'
mysql -uroot -e 'CREATE DATABASE isu6fportal;'
mysql -uroot -Disu6fportal < db/schema.sql
cat data/teams.tsv | go run cmd/importteams/main.go
```

これでチームデータと運営のデータが入るので、以下のコマンドでポータルを起動。

```
make
./portal -database-dsn="root:@/isu6fportal"
```

