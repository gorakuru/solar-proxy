# solar-proxy

建て得で使われているsolar-minitorのdigest認証をキャンセルするためのプロキシです。
ProxmoxのLXCでデプロイする想定ですが、goなのでどこでも動くと思います。



# デプロイ方法

## 1. LXCコンテナを用意

Ubunut 26で確認ずみ

下記のパッケージを導入しておく
```bash
apt install -y golang git
```


## 2. ソースのダウンロード

```bash
cd /opt
git clone https://github.com/gorakuru/solar-proxy.git
cd solar-proxy
```

## 3. ビルド

```bash
go mod init solar-proxy
go mod tidy
go build -o solar-proxy main.go
```

## 4. サービスファイルの用意

### .envファイルの修正

.envファイルを開き、環境に合わせてパスワードなどを修正する


### serviceファイルの配置

```bash
ln -s /opt/solar-proxy/solar-proxy.service /etc/systemd/system/solar-proxy.service
systemctl daemon-reload
```

## 5. 自動起動の設定と起動

```bash
systemctl enable solar-proxy
systemctl start solar-proxy
```

