# Chouseisan Reminder
[![Build Status](https://travis-ci.org/nowsprinting/ChouseisanReminder.svg?branch=master)](https://travis-ci.org/nowsprinting/ChouseisanReminder)
[![Coverage Status](https://coveralls.io/repos/github/nowsprinting/ChouseisanReminder/badge.svg)](https://coveralls.io/github/nowsprinting/ChouseisanReminder)

出欠管理ツール「[調整さん](https://chouseisan.com/)」から出欠情報を取得し、期日（開催日）の数日前に参加状況を流すLINE BOT


## 動作仕様

### LINEからのコールバック受信時

#### 友だち登録イベント

- 対象のMIDを購読者としてデータストアに追加
- 購読者全員に「購読を開始しました」メッセージを送信

#### ブロック（友だち登録解除）イベント

- 対象のMIDを購読者から削除
- 購読者全員に「購読を解除しました」メッセージを送信

#### トーク受信

- 購読者全員に同様のテキストを送信する。このとき送信者のスクリーンネームも添える
- その他、直近の参加予定を返すなどのコマンド。追々検討

#### スタンプ受信

- 購読者全員に同様のスタンプを送信する。このとき送信者のスクリーンネームも添える

### 定時実行

- 毎日8:00に定時実行
- 3日後、および、当日に調整さんの予定があれば、購読者全員に出欠入力状況を送信

### Webブラウザからのアクセス時

- usage.htmlを表示する


## 動作環境

- Google App Engine SDK for Go 1.9.40
- Go SDK for the LINE Messaging API


## 設定ファイル

以下のファイルを準備する

### app.yaml

`app.yaml`にはLINE BOTのキー情報などを含むため、リポジトリから除外している。下記の書式で`app.yaml`を作成すること

	application: YOUR_APPLICATION_ID
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
	  CHOUSEISAN_EVENT_HASH: 'YOUR_EVENT_HASH'
	  LINE_CHANNEL_SECRET: 'YOUR_CHANNEL_SECRET'
      LINE_CHANNEL_ACCESS_TOKEN: `YOUR_ACCESS_TOKEN`

### LINE BOTのQRコード

LINE BOTのQRコードを`/img/linebot_qr.png`に置くこと（usage.htmlからリンクしている）


## LINE BOTについて

- トライアルで提供されていた BOT API Trialはdeprecatedされたため、新しいMessaging APIを利用するよう書き換えた。
    - see: https://developers.line.me/messaging-api/overview
- 開設したチャンネルの"Basic Information"にある"Callback URL"に、コールバックを受け取るURLを設定する必要がある。以下のように設定する。
	- https:// YOUR-PROJECT-ID .appspot.com:443/line/callback
- 新しいMessaging APIではBOTをグループに追加できるが、本BOTは従来のBOT APIベースのため、リマインダを受けたいユーザは個々にBOTを友だち登録して利用する形式を取る。


## 調整さんについて

- https://chouseisan.com/
- 開催日決定ツールとしてでなく、出欠確認ツールとして利用する
