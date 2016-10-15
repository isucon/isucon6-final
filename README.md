# ISUCON6 final

## イメージのデプロイ

### portal

<a href="https://portal.azure.com/#create/Microsoft.Template/uri/https%3A%2F%2Fgist.githubusercontent.com%2Fcatatsuy%2F231661330222701bf1ca976d9412d8df%2Fraw%2F135bd621c4164ca1436620b425f71d6516728c3b%2Fportal.json" target="_blank">
  <img src="http://azuredeploy.net/deploybutton.png"/>
</a>

### proxy

`provisioning/external_vars.yml`の`portal_private_ip`をportalのprivate ipにする。これをしないとセットアップに失敗する。

`provisioning/proxy`ディレクトリ以下で、以下のような`prodution`というファイル名のものを作る。`~/.ssh/config`を作る必要があるのと、__azure上でVMを作る際に名前にproxyという名前が含まれている必要がある（重要）__。

```
[build_servers]
isu6f-proxy01
isu6f-proxy02
isu6f-proxy03
```

以下のコマンドを実行する。

```
cd provisioning/proxy
ansible-playbook -i production playbook/*.yml
```

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
