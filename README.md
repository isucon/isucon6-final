# ISUCON6 final

## イメージのデプロイ

### portal

<a href="https://portal.azure.com/#create/Microsoft.Template/uri/https%3A%2F%2Fgist.githubusercontent.com%2Fcatatsuy%2F231661330222701bf1ca976d9412d8df%2Fraw%2F135bd621c4164ca1436620b425f71d6516728c3b%2Fportal.json" target="_blank">
  <img src="http://azuredeploy.net/deploybutton.png"/>
</a>

### proxy

`ansible/playbook/roles/proxy/vars/main.yml.tmp`のtmpを取り、portalのprivate ipを記述する。これをしないとセットアップに失敗する。

`ansible`ディレクトリ以下で、以下のような`prodution`というファイル名のものを作る。`~/.ssh/config`を作る必要があるのと、__azure上でVMを作る際に名前にproxyという名前が含まれている必要がある（重要）__。

```
[build_servers]
isu6f-proxy01
isu6f-proxy02
isu6f-proxy03
```

以下のコマンドを実行する。

```
ansible-playbook -i production playbook/setup.yml --tags=proxy
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
