# ISUCON6 final

## イメージのデプロイ

### portal

最初に起動する必要がある。

  * [/provisioning/portal/deploy.sh](/provisioning/portal/deploy.sh)を適切な環境変数を渡して実行
    * VM名にはportalと含む
  * portalディレクトリ以下で `make portal_linux_amd64` `make importteams_linux_amd64` をする
  * 立てたサーバーにsshできるように`~/.ssh/config`に書く
  * [/provisioning/portal/](/provisioning/portal/)ディレクトリ以下で`production`というファイル名で、`~/.ssh/config`に設定したホスト名を書いておく
  * `ansible-playbook -i production ansible/*.yml`を実行する

### bench

  * `provisioning/external_vars.yml`の`portal_private_ip`をportalのprivate ipにする
    * VM名にはbenchと含む
  * [/provisioning/bench/deploy.sh](/provisioning/bench/deploy.sh)を適切な環境変数を渡して実行
  * benchディレクトリ以下で `make isucon6f` をする
  * 立てたサーバーにsshできるように`~/.ssh/config`に書く
  * [/provisioning/bench/](/provisioning/bench/)ディレクトリ以下で`production`というファイル名で、`~/.ssh/config`に設定したホスト名を書いておく
  * `ansible-playbook -i production ansible/*.yml`を実行する

### proxy

  * `provisioning/external_vars.yml`の`portal_private_ip`をportalのprivate ipにする
    * VM名にはproxyと含む（必須）
  * [/provisioning/proxy/deploy.sh](/provisioning/proxy/deploy.sh)を適切な環境変数を渡して実行

proxyは以下のような挙動をする。

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
