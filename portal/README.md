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

終了一時間前あたりで `team_scores_snapshot` テーブルを作るとリーダーボードが固定されます。

    INSERT INTO team_scores_snapshot SELECT * FROM team_scores

## 開発・運用むけ情報

秘密のURLです。認証とかはとくになし

- http://isucon6q.catatsuy.org/mBGWHqBVEjUSKpBF/debug/vars
- http://isucon6q.catatsuy.org/mBGWHqBVEjUSKpBF/debug/queue
- http://isucon6q.catatsuy.org/mBGWHqBVEjUSKpBF/debug/leaderboard

## ローカルで開発する

```
mysql -uroot -e 'DROP DATABASE IF EXISTS isu6fportal_day0;'
mysql -uroot -e 'CREATE DATABASE isu6fportal_day0;'
mysql -uroot -Disu6fportal_day0 < db/schema.sql
mysql -uroot -Disu6fportal_day1 < db/schema.sql
cat data/teams.tsv | go run cmd/importteams/main.go -dsn-base="root:@"
```

これでチームデータと運営のデータが入るので、以下のコマンドでポータルを起動。

```
make
./portal -database-dsn="root:@/isu6fportal_day0"
```

テストは上記のコマンドでMySQLを初期化した後に、

```
go test
```

する。（テスト内で初期化してくれたりはしない）
