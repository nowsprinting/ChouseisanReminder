package main

import (
	"encoding/csv"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"

	"golang.org/x/net/context"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

// 調整さんの開催日ごとの集計エントリ
type schedule struct {
	Date             time.Time // 開催日（時間は00:00:00JST）
	DateString       string    // 日程欄（文字列）
	Present          int       // ◯
	Absent           int       // ×
	Unknown          int       // △および未入力
	ParticipantsName string    // 参加者の名前を列挙したもの
	UnknownName      string    // △および未入力の名前を列挙したもの
}

// 送信メッセージ用のサマリを組み立てて返す
func (s *schedule) constructSummaryBody() string {
	return s.DateString + "の出欠状況をお知らせします\n\n" +
		"参加: " + strconv.Itoa(s.Present) + "名" + s.ParticipantsName +
		"\n不参加: " + strconv.Itoa(s.Absent) + "名\n" +
		"不明/未入力: " + strconv.Itoa(s.Unknown) + "名" + s.UnknownName
}

// 送信メッセージ用のサマリを組み立てて返す
func (s *schedule) constructSummary(hash string) string {
	return s.constructSummaryBody() +
		"\n\n詳細および出欠変更は「調整さん」へ\nhttps://chouseisan.com/s?h=" + hash
}

// 調整さんスケジュールのMap型
type scheduleMap map[string]schedule

/**
 * 調整さんcsvをパースして、参加人数などを集計する
 */
func parseCsv(c context.Context, csvBody io.ReadCloser, today time.Time) (m scheduleMap) {
	var (
		names    []string
		rowCount = 0
	)
	m = make(scheduleMap)

	reader := csv.NewReader(transform.NewReader(csvBody, japanese.ShiftJIS.NewDecoder()))
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if e2, ok := err.(*csv.ParseError); ok && e2.Err == csv.ErrFieldCount {
			//フィールド数エラーは無視
		} else if err != nil {
			log.Errorf(c, "Read chouseisan's csv failed. err: %v", err)
			return nil
		}

		if rowCount < 2 {
			//イベント名、詳細説明文はスキップ

		} else if rowCount == 2 {
			//名前行
			for i, v := range row {
				if i > 0 {
					names = append(names, v)
				}
			}

		} else {
			//データ行（最終のコメント行も含む）
			s := schedule{}
			for i, v := range row {
				if i == 0 {
					//日付カラムはパースしてキーにする
					tz, _ := time.LoadLocation("Asia/Tokyo")
					year := today.Year()
					r := regexp.MustCompile(`^(\d{1,2})/(\d{1,2}).*$`)
					md := r.FindAllStringSubmatch(v, -1)
					if len(md) == 0 {
						log.Debugf(c, "Month and day parse error. col:%v", v)
						continue
					}
					month, merr := strconv.Atoi(md[0][1])
					day, derr := strconv.Atoi(md[0][2])
					if merr != nil {
						log.Debugf(c, "Month parse error. col:%v error:%v", v, merr)
						continue
					} else if derr != nil {
						log.Debugf(c, "Month parse error. col:%v error:%v", v, merr)
						continue
					} else {
						//日付パース成功
						if time.Month(month) < today.Month() {
							year++ //来年として扱う
						}
						s.Date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, tz)
						s.DateString = v
						s.Present = 0
						s.Absent = 0
						s.Unknown = 0
						s.ParticipantsName = ""
						s.UnknownName = ""
					}

				} else if len(names[i-1]) > 0 {
					//出欠カラムの内容を、scheduleに足しこむ
					if v == "○" {
						s.Present++
						if len(s.ParticipantsName) > 0 {
							s.ParticipantsName += ","
						}
						s.ParticipantsName += names[i-1]
					} else if v == "×" {
						s.Absent++
					} else {
						s.Unknown++
						if len(s.UnknownName) > 0 {
							s.UnknownName += ","
						}
						s.UnknownName += names[i-1]
					}
				}
			}
			if len(s.DateString) > 0 {
				if len(s.ParticipantsName) > 0 {
					s.ParticipantsName = "(" + s.ParticipantsName + ")"
				}
				if len(s.UnknownName) > 0 {
					s.UnknownName = "(" + s.UnknownName + ")"
				}
				m[s.Date.String()] = s
			}
		}
		rowCount++
	}
	return m
}

/**
 * 購読者ごとのイテレーション処理。調整さんをクロールして通知対象があれば集計して返す
 */
func chouseisanIterator(current *subscriber, c context.Context, client *http.Client, w http.ResponseWriter, r *http.Request) []schedule {
	result := []schedule{}

	//調整さんの"出欠表をダウンロード"リンクからcsv形式で取得
	url := "https://chouseisan.com/schedule/List/createCsv?h=" + current.ChouseisanHash
	res, err := client.Get(url)
	if err != nil {
		log.Errorf(c, "Get chouseisan's csv failed. err: %v", err)
		return result
	} else if res.StatusCode != 200 {
		log.Errorf(c, "Get chouseisan's csv failed. StatusCode: %v", res.StatusCode)
		return result
	}

	//csvをパース
	tz, _ := time.LoadLocation("Asia/Tokyo")
	today := time.Now().In(tz)
	m := parseCsv(c, res.Body, today)

	//当日の予定をピック
	targetDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, tz)
	obj, exist := m[targetDate.String()]
	if exist {
		result = append(result, obj)
	} else {
		log.Debugf(c, "Not found schedule at today.")
	}

	//3日後の予定をピック
	targetDate = targetDate.AddDate(0, 0, 3)
	obj, exist = m[targetDate.String()]
	if exist {
		result = append(result, obj)
	} else {
		log.Debugf(c, "Not found schedule at 3 days after.")
	}

	return result
}

/**
 * 調整さんをクロールして出欠を通知
 *
 * 引数にContextとhttp.Clientを取るインナーメソッド
 */
func crawlChouseisanWithContext(c context.Context, client *http.Client, w http.ResponseWriter, r *http.Request) {
	bot, err := createBotClient(c, client)
	if err != nil {
		return
	}

	//hashの入ってるエンティティを抽出してループ
	ite := datastore.NewQuery("Subscriber").Run(c)
	for {
		var cSubscriber subscriber
		_, err = ite.Next(&cSubscriber)
		if err == datastore.Done {
			break
		} else if err != nil {
			log.Errorf(c, "Error occurred at fetch Subscriber. err:%v", err)
			break
		}

		if cSubscriber.ChouseisanHash != "" {
			// ハッシュが設定されていれば、調整さんイベントをクロール
			log.Infof(c, "Crawl chouseisan! subscriber:%v hash:%v", cSubscriber.DisplayName, cSubscriber.ChouseisanHash)
			result := chouseisanIterator(&cSubscriber, c, client, w, r)

			// リマインド対象イベントがあれば、Push Messageを送信
			for _, v := range result {
				log.Infof(c, "Remind event! subscriber:%v date:%v", cSubscriber.DisplayName, v.DateString)
				template := linebot.NewButtonsTemplate(
					"", //サムネイル
					"", //タイトル
					v.constructSummaryBody(), //画像もタイトルも指定しない場合：160文字以内
					linebot.NewURITemplateAction("出欠を登録（変更）する", "https://chouseisan.com/s?h="+cSubscriber.ChouseisanHash),
				)
				altText := "このメッセージが見えている人は、お使いのLINEアプリのバージョンおよび機種名を教えてください"
				if _, err = bot.PushMessage(cSubscriber.MID, linebot.NewTemplateMessage(altText, template)).Do(); err != nil {
					log.Errorf(c, "Error occurred at crawl chouseisan. subscriber:%v, date:%v, err: %v", cSubscriber.DisplayName, v.DateString, err)
				}
			}
		}
	}
}

/**
 * 調整さんをクロールして出欠を通知（cronからキックされる）
 */
func crawlChouseisan(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	crawlChouseisanWithContext(c, urlfetch.Client(c), w, r)
}
