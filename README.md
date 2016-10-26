# Chouseisan Reminder Bot
[![Build Status](https://travis-ci.org/nowsprinting/ChouseisanReminder.svg?branch=master)](https://travis-ci.org/nowsprinting/ChouseisanReminder)
[![Coverage Status](https://coveralls.io/repos/github/nowsprinting/ChouseisanReminder/badge.svg)](https://coveralls.io/github/nowsprinting/ChouseisanReminder)

出欠管理ツール「[調整さん](https://chouseisan.com/)」から出欠情報を取得し、期日（開催日）の数日前に参加状況を流すLINE BOT


## 動作仕様

### LINEからのコールバック受信時

##### 友だち登録、グループ/ルームへの招待イベント

- 対象のIDを購読者としてデータストアに追加
- 送信者（ユーザ/グループ/ルーム）に「リマインダを登録しました」メッセージを送信

##### ブロック（友だち登録解除）、グループからの削除イベント

- 対象のIDを購読者から削除
- メッセージは送信しない（受け取る相手がいないため）
- トークルームからのleaveイベントもグループと同様の処理をするが、実際にはこのイベントは送信されない

##### トーク受信

- `/set chouseisan`コマンドで、リマインド対象の調整さんイベントを設定できる
- グループ利用を想定しているため、テキストメッセージのオウム返しはしない

### 定時実行

- 毎日8:00に定時実行し、購読者ごとの調整さんイベント日程をクロール
- 3日後もしくは当日の予定があれば、その購読者に出欠入力状況を送信
	- ここで`Push Message`APIを使用するため、BOTアカウントの契約プランはDeveloper Trialかプロ以上が必要。

### Webブラウザからのアクセス時

- usage.htmlを表示する


## 動作環境

- Google App Engine SDK for Go 1.9.40
- Go SDK for the LINE Messaging API


## 設定ファイル

以下のファイルを準備する

### app.yaml

`app.yaml`にはLINE BOTのキー情報などを含むため、リポジトリから除外している。下記の書式で`app.yaml`を作成すること

	application: YOUR-APPLICATION-ID
	version: 2
	runtime: go
	api_version: go1

	handlers:
	- url: /img
	  static_dir: img
	- url: /task/.*
	  script: _go_app
	  login: admin
	- url: /cron/.*
	  script: _go_app
	  login: admin
	- url: /.*
	  script: _go_app

	env_variables:
	  LINE_CHANNEL_SECRET: 'YOUR_CHANNEL_SECRET'
	  LINE_CHANNEL_ACCESS_TOKEN: 'YOUR_ACCESS_TOKEN'

### LINE BOTのQRコード

LINE BOTのQRコードを`/img/linebot_qr.png`に置くこと（usage.htmlからリンクしている）


## LINE BOTについて

- トライアルで提供されていた BOT API Trialはdeprecatedされたため、新しいMessaging APIを利用するよう書き換えた
    - see: [LINE Developers - Messaging API - Overview](https://developers.line.me/messaging-api/overview)
- 開設したチャンネルの"Basic Information"にある"Callback URL"に、コールバックを受け取るURLを設定する必要がある。記述例:
	- `https:// YOUR-APPLICATION-ID .appspot.com:443/line/callback`
- その他、LINE BOTまわりは下記ブログエントリを参照
	- [調整さんリマインダLINE BOTを作ってみた - やらなイカ？](http://nowsprinting.hatenablog.com/entry/2016/08/23/000000)
	- [LINEの新しいMessaging APIを試してみた - やらなイカ？](http://nowsprinting.hatenablog.com/entry/2016/10/02/043410)


## 調整さんについて

- https://chouseisan.com/
- 開催日決定ツールとしてでなく、出欠確認ツールとして利用する
