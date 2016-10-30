# ISUCON6 final

[当日のレギュレーション](/regulation.md)も参照してください。

## ディレクトリ構成

```
├── ansible      # 競技者用インスタンスのセットアップ用ansble
├── azure        # 競技者用インスタンスのdeploy to Azure
├── bench        # ベンチマーカーのソースコード
├── portal       # ポータルサイトのソースコード
├── provisioning # 運営側（portal/bench/proxy）のセットアップ用ansible
└── webapp       # 各言語の参考実装
```

## 競技者用アプリケーション

<a href="https://portal.azure.com/#create/Microsoft.Template/uri/https%3A%2F%2Fgithub.com%2Fisucon%2Fisucon6-final%2Fraw%2Fmaster%2Fazure%2Fazuredeploy.json" target="_blank"><img src="http://azuredeploy.net/deploybutton.png"/></a>
<a href="http://armviz.io/#/?load=https%3A%2F%2Fraw.githubusercontent.com%2Fmatsuu%2Fisucon6-final%2Ffeature%2Fupdate-readme%2Fazure%2Fazuredeploy.json" target="_blank"><img src="http://armviz.io/visualizebutton.png"/></a>

DockerだけインストールされていればOS等は問わない。

webapp 以下を競技用マシンの /home/isucon/webapp に置けばセットアップ完成。

起動方法などは [webapp/README.md](/webapp/README.md) に書いた。

## portal, bench, proxy

<a href="https://portal.azure.com/#create/Microsoft.Template/uri/https%3A%2F%2Fgithub.com%2Fisucon%2Fisucon6-final%2Fraw%2Fmaster%2Fprovisioning%2Fdeploy.json" target="_blank"><img src="http://azuredeploy.net/deploybutton.png"/></a>
<a href="http://armviz.io/#/?load=https%3A%2F%2Fraw.githubusercontent.com%2Fmatsuu%2Fisucon6-final%2Ffeature%2Fupdate-readme%2Fprovisioning%2Fdeploy.json" target="_blank"><img src="http://armviz.io/visualizebutton.png"/></a>

portalに試しにログインしたい

  * 一般アカウント
    * user: 1 pass: y8aaZLdAXAXn
  * 運営アカウント
    * user: 9999 pass: Btw5R5fskVvXOzT

編集する場合は、[/portal/cmd/importteams/main.go](/portal/cmd/importteams/main.go) と [/portal/data/teams.tsv](/portal/data/teams.tsv) を参照のこと

### portal

最初に起動する必要がある。consulが起動している。

### bench

最低1台起動している必要がある。

### proxy

VM名にはproxyと含む必要がある。consulが動くために最低3台必要なので、proxyは2台以上起動する必要がある。proxyは以下のような挙動をする。

  * portalと同じconsulのクラスタの一員になる
  * proxyはセットアップ時に、portalからnginxの設定を取得して設定を反映する
  * portalは参加者のIPアドレスが変更されたら、`consul event -name="nginx_reload" -node="proxy"`を叩く
    * `/usr/local/bin/nginx_reload`を叩く
    * 全proxyはportalからnginxの設定を取得して設定を反映する
  * proxyは投入されると、consulのクラスタにjoinする
    * `/usr/local/bin/update_members`を叩く
    * portalはmemberが増えたら、portalのDBに登録して、ベンチマーカーに渡す

proxyを減らす場合は以下の手順が必要

  * 減らすインスタンス上で`consul leave`と打つ

#### proxy全台をconsulで操作したい

  * proxy全台のnginxを起動したい
    * `consul exec -node proxy "sudo systemctl start nginx"`
  * consulのイベントを発行する
    * `consul event -name="nginx_reload" -node="proxy"`

### Azure-CLIを使う方法

#### install

```
npm install azure-cli -g
```

#### login

```
azure login
azure account list
azure account set <ID>
```

#### Portalデプロイ例

##### parameters.json

```json
{
    "vmName":{
        "value": "isucon6f-ex-portal(change name if you want)"
    },
    "sshPublicKey": {
        "value": "ssh-rsa ...(your ssh-public-key)"
    }
}
```

portalが要求する変数を書いておく。

##### command

```
azure group deployment create --template-file deploy.json -e parameters.json isucon6-final-dev
```

* -e
  * パラメータ設定用ファイルの指定
* --template-file
  * テンプレートファイルの指定
