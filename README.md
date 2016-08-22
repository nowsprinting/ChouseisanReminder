# Chouseisan Reminder

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

### 定時実行

- 毎日8:00に定時実行
- 3日後に調整さんの予定があれば、購読者全員に出欠入力状況を送信

### Webブラウザからのアクセス時

- usage.htmlを表示する


## 動作環境

- Google App Engine SDK for Go 1.9.40
- LINE BOT API Trial SDK/GO Version


## 設定ファイル

以下のファイルを準備する

### app.yaml

`app.yaml`にはLINE BOTのキー情報などを含むため、リポジトリから除外している。下記の書籍で`app.yaml`を作成すること

	application: YOUR_APPLICATION_ID
	version: 1
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
	  LINE_CHANNEL_ID: 'YOUR_CHANNEL_ID'
	  LINE_CHANNEL_SECRET: 'YOUR_CHANNEL_SECRET'
	  LINE_CHANNEL_MID: 'YOUR_CHANNEL_MID'

### LINE BOTのQRコード

LINE BOTのQRコードを`/img/linebot_qr.png`に置くこと（usage.htmlからリンクしている）


## LINE BOTについて

- 現状はトライアル（BOT API Trial Account）
	- https://business.line.me/services/products/4/introduction
- 開設したチャンネルの"Basic Information"にある"Callback URL"に、コールバックを受け取るURLを設定
	- https://YOUR_PROJECT_ID.appspot.com:443/line/callback
- 現状、BOTはグループに追加することができない。そのためリマインダを受けたいユーザは個々にBOTを友だち登録する必要がある
- BOT API Trial Accountでは、友だち登録は50人まで。制限解除を申請すれば5,000人に拡張できる（要審査）


## 調整さんについて

- https://chouseisan.com/
- 開催日決定ツールとしてでなく、出欠確認ツールとして利用する
