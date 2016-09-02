package main

import (
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"google.golang.org/appengine"
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
}

// 調整さんスケジュールのMap型
type scheduleMap map[string]*schedule

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
			s := new(schedule)
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
					}
				}
			}
			if len(s.DateString) > 0 {
				m[s.Date.String()] = s
			}
		}
		rowCount++
	}
	return m
}

/**
 * 調整さんをクロールして出欠を通知（cronからキックされる）
 */
func crawlChouseisan(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	//調整さんの"出欠表をダウンロード"リンクからcsv形式で取得
	url := "https://chouseisan.com/schedule/List/createCsv?h=" + os.Getenv("CHOUSEISAN_EVENT_HASH")
	client := urlfetch.Client(c)
	res, err := client.Get(url)
	if err != nil {
		log.Errorf(c, "Get chouseisan's csv failed. err: %v", err)
		return
	} else if res.StatusCode != 200 {
		log.Errorf(c, "Get chouseisan's csv failed. StatusCode: %v", res.StatusCode)
		return
	}

	//csvをパース
	tz, _ := time.LoadLocation("Asia/Tokyo")
	today := time.Now().In(tz)
	m := parseCsv(c, res.Body, today)

	//当日の予定をピック
	targetDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, tz)
	obj, exist := m[targetDate.String()]
	if exist {
		sendSchedule(c, obj)
	} else {
		log.Infof(c, "Not found schedule at today.")
	}

	//3日後の予定をピック
	targetDate = targetDate.AddDate(0, 0, 3)
	obj, exist = m[targetDate.String()]
	if exist {
		sendSchedule(c, obj)
	} else {
		log.Infof(c, "Not found schedule at 3 days after.")
	}
}

/**
 * 出欠メッセージを組み立ててLINEに送信
 */
func sendSchedule(c context.Context, obj *schedule) {
	//メッセージを組み立てて送信
	bot, err := createBotClient(c)
	if err != nil {
		return
	}
	msg := obj.DateString + "の出欠状況をお知らせします\n参加: " + strconv.Itoa(obj.Present) + "名(" + obj.ParticipantsName + ")\n不参加: " + strconv.Itoa(obj.Absent) + "名\n不明/未入力: " + strconv.Itoa(obj.Unknown) + "名\n\n詳細および出欠変更は「調整さん」へ\nhttps://chouseisan.com/s?h=" + os.Getenv("CHOUSEISAN_EVENT_HASH")
	sendToAll(c, bot, msg)
	return
}
