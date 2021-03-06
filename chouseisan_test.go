package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/thingful/httpmock"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
)

/**
 * リマインド通知用のサマリ組み立てのテスト
 */
func TestConstructSummary(t *testing.T) {
	testdata := schedule{
		DateString:       "10/29(土)",
		Present:          1,
		Absent:           2,
		Unknown:          4,
		ParticipantsName: "(電二郎)",
		UnknownName:      "(電一,電四郎,電五郎,電六郎)",
	}
	expectedSummary := "10/29(土)の出欠状況をお知らせします\n\n" +
		"参加: 1名(電二郎)\n不参加: 2名\n不明/未入力: 4名(電一,電四郎,電五郎,電六郎)" +
		"\n\n詳細および出欠変更は「調整さん」へ\n" +
		"https://chouseisan.com/s?h=3f7ffd73ba174332ae05bd363eba8e71"
	actualSummary := testdata.constructSummary("3f7ffd73ba174332ae05bd363eba8e71")
	if actualSummary != expectedSummary {
		t.Errorf("Unmatch summary\nexpect:\n%v\nactual:\n%v", expectedSummary, actualSummary)
	}
}

/**
 * 正常ケース
 */
func TestParseCsvNormally(t *testing.T) {
	var (
		objDay time.Time
		obj    schedule
		exist  bool
		err    error
	)
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/chouseisan/normally.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer testdata.Close()

	tz, _ := time.LoadLocation("Asia/Tokyo")
	today := time.Date(2016, time.December, 1, 0, 0, 0, 0, tz)
	m := parseCsv(c, testdata, today)

	//12.24
	objDay = time.Date(2016, time.December, 24, 0, 0, 0, 0, tz)
	obj, exist = m[objDay.String()]
	if !exist {
		t.Fatalf("Entry not found. date is %v", objDay)
	}
	if obj.Present != 4 {
		t.Errorf("Bad obj.Present: %v", obj.Present)
	}
	if obj.Absent != 1 {
		t.Errorf("Bad obj.Absent: %v", obj.Absent)
	}
	if obj.Unknown != 2 {
		t.Errorf("Bad obj.Unknown: %v", obj.Unknown)
	}
	if obj.ParticipantsName != "(電三太郎,電四郎,電六郎,電七郎)" {
		t.Errorf("Bad obj.ParticipantsName: %v", obj.ParticipantsName)
	}
	if obj.UnknownName != "(電一,電五郎)" {
		t.Errorf("Bad obj.UnknownName: %v", obj.UnknownName)
	}
}

/**
 * 翌年扱いになる日程のテスト
 */
func TestParseCsvNormallyNextYear(t *testing.T) {
	var (
		objDay time.Time
		obj    schedule
		exist  bool
		err    error
	)
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/chouseisan/normally.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer testdata.Close()

	tz, _ := time.LoadLocation("Asia/Tokyo")
	today := time.Date(2016, time.December, 1, 0, 0, 0, 0, tz)
	m := parseCsv(c, testdata, today)

	//11月は翌年扱い
	objDay = time.Date(2017, time.November, 26, 0, 0, 0, 0, tz)
	obj, exist = m[objDay.String()]
	if !exist {
		t.Fatalf("Entry not found. date is %v", objDay)
	}
	if obj.Present != 5 {
		t.Errorf("Bad obj.Present: %v", obj.Present)
	}
	if obj.Absent != 1 {
		t.Errorf("Bad obj.Absent: %v", obj.Absent)
	}
	if obj.Unknown != 1 {
		t.Errorf("Bad obj.Unknown: %v", obj.Unknown)
	}
	if obj.ParticipantsName != "(電一,電次郎,電四郎,電六郎,電七郎)" {
		t.Errorf("Bad obj.ParticipantsName: %v", obj.ParticipantsName)
	}
	if obj.UnknownName != "(電五郎)" {
		t.Errorf("Bad obj.UnknownName: %v", obj.UnknownName)
	}
}

/**
 * カラムの無いケース
 */
func TestParseCsvNormallyNoCol(t *testing.T) {
	var (
		objDay time.Time
		obj    schedule
		exist  bool
		err    error
	)
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/chouseisan/no_col.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer testdata.Close()

	tz, _ := time.LoadLocation("Asia/Tokyo")
	today := time.Date(2016, time.December, 1, 0, 0, 0, 0, tz)
	m := parseCsv(c, testdata, today)

	//12.24
	objDay = time.Date(2016, time.December, 17, 0, 0, 0, 0, tz)
	obj, exist = m[objDay.String()]
	if !exist {
		t.Fatalf("Entry not found. date is %v", objDay)
	}
	if obj.Present != 0 {
		t.Errorf("Bad obj.Present: %v", obj.Present)
	}
	if obj.Absent != 0 {
		t.Errorf("Bad obj.Absent: %v", obj.Absent)
	}
	if obj.Unknown != 0 {
		t.Errorf("Bad obj.Unknown: %v", obj.Unknown)
	}
	if obj.ParticipantsName != "" {
		t.Errorf("Bad obj.ParticipantsName: %v", obj.ParticipantsName)
	}
	if obj.UnknownName != "" {
		t.Errorf("Bad obj.UnknownName: %v", obj.UnknownName)
	}
}

/**
 * 行の無いケース
 */
func TestParseCsvNormallyNoRow(t *testing.T) {
	var (
		objDay time.Time
		obj    schedule
		exist  bool
		err    error
	)
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/chouseisan/no_row.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer testdata.Close()

	tz, _ := time.LoadLocation("Asia/Tokyo")
	today := time.Date(2016, time.December, 1, 0, 0, 0, 0, tz)
	m := parseCsv(c, testdata, today)

	//12.24（no_rowには存在しない）
	objDay = time.Date(2016, time.December, 17, 0, 0, 0, 0, tz)
	obj, exist = m[objDay.String()]
	if exist {
		t.Errorf("Entry found??? date is %v, find date is %v", objDay, obj.DateString)
	}
}

/**
 * 日程カラムのフォーマット不正データ
 */
func TestParseCsvInvalidDateFormat(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/chouseisan/invalid.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer testdata.Close()

	tz, _ := time.LoadLocation("Asia/Tokyo")
	today := time.Date(2016, time.December, 1, 0, 0, 0, 0, tz)
	m := parseCsv(c, testdata, today)

	//パース結果は4件であること（不正フォーマットはスキップされ、日付の0日や32日はそれなりに解釈されていること）
	if len(m) != 4 {
		t.Errorf("Invalid parse result.\n%v", m)
	}
}

/**
 * 調整さんクロール処理のテスト（正常系）
 */
func TestCrawlChouseisan(t *testing.T) {
	opt := aetest.Options{StronglyConsistentDatastore: true} //データストアに即反映
	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatalf("Failed to create aetest instance: %v", err)
	}
	defer instance.Close()

	// http.Requestを生成
	req, err := instance.NewRequest("POST", "/cron/crawlchouseisan", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") //必須

	// Contextとhttp.Clientは、テストコード側でインスタンス化する（モックと共通のインスタンスを使う必要があるため）
	ctx := appengine.NewContext(req)
	client := urlfetch.Client(ctx)

	// 調整さんへのリクエストをモックする
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"https://chouseisan.com/schedule/List/createCsv?h=3f7ffd73ba174332ae05bd363eba8e71",
			httpmock.NewStringResponder(200, readFile(t, "testdata/chouseisan/normally.csv")),
		),
	)
	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"https://chouseisan.com/schedule/List/createCsv?h=11111111111111111111111111111111",
			httpmock.NewStringResponder(200, readFile(t, "testdata/chouseisan/normally.csv")),
		),
	)
	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"https://chouseisan.com/schedule/List/createCsv?h=22222222222222222222222222222222",
			httpmock.NewStringResponder(200, readFile(t, "testdata/chouseisan/normally.csv")),
		),
	)

	// 購読者エンティティを用意しておく
	entities := []subscriber{
		{
			MID:            "C00000000000000000000000000000000",
			ChouseisanHash: "3f7ffd73ba174332ae05bd363eba8e71",
		}, {
			MID:            "R00000000000000000000000000000001",
			ChouseisanHash: "11111111111111111111111111111111",
		}, {
			MID:            "U00000000000000000000000000000002",
			ChouseisanHash: "22222222222222222222222222222222",
		}}
	for _, current := range entities {
		key := datastore.NewKey(ctx, "Subscriber", current.MID, 0, nil)
		if _, err = datastore.Put(ctx, key, &current); err != nil {
			t.Fatal(err)
		}
	}

	// execute
	res := httptest.NewRecorder()
	crawlChouseisanWithContext(ctx, client, res, req) //モックと同じhttp.Clientインスタンスを渡す

	// リクエストは正常終了していること
	if res.Code != http.StatusOK {
		t.Errorf("Non-expected status code: %v\n\tbody: %v", res.Code, res.Body)
	}

	// スタブがすべて呼ばれたことを検証
	if err := httpmock.AllStubsCalled(); err != nil {
		t.Errorf("Not all stubs were called: %s", err)
	}
}

/**
 * 調整さんクロール処理のテスト（対象の購読者エンティティなし）
 */
func TestCrawlChouseisanZeroSubscriber(t *testing.T) {
	opt := aetest.Options{StronglyConsistentDatastore: true} //データストアに即反映
	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatalf("Failed to create aetest instance: %v", err)
	}
	defer instance.Close()

	// http.Requestを生成
	req, err := instance.NewRequest("POST", "/cron/crawlchouseisan", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") //必須

	// Contextとhttp.Clientは、テストコード側でインスタンス化する（モックと共通のインスタンスを使う必要があるため）
	ctx := appengine.NewContext(req)
	client := urlfetch.Client(ctx)

	// 調整さんへのリクエストをモックする（呼ばれることはないが、検証のため）
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	// 購読者エンティティは1件だが、調整さんハッシュは持っていない
	entity := subscriber{
		MID:            "C00000000000000000000000000000000",
		ChouseisanHash: "",
	}
	key := datastore.NewKey(ctx, "Subscriber", "C00000000000000000000000000000000", 0, nil)
	if _, err = datastore.Put(ctx, key, &entity); err != nil {
		t.Fatal(err)
	}

	// execute
	res := httptest.NewRecorder()
	crawlChouseisanWithContext(ctx, client, res, req) //モックと同じhttp.Clientインスタンスを渡す

	// リクエストは正常終了していること
	if res.Code != http.StatusOK {
		t.Errorf("Non-expected status code: %v\n\tbody: %v", res.Code, res.Body)
	}

	// スタブがすべて呼ばれたことを検証
	if err := httpmock.AllStubsCalled(); err != nil {
		t.Errorf("Not all stubs were called: %s", err)
	}
}
