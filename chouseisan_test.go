package main

import (
	"os"
	"testing"
	"time"

	"google.golang.org/appengine/aetest"
)

/*
 * 正常ケース
 */
func TestParseCsv_normally(t *testing.T) {
	var (
		objDay time.Time
		obj    *schedule
		exist  bool
		err    error
	)
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/chouseisan.csv")
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
		t.Errorf("Entry not found. date is %v", objDay)
		return
	}
	if obj.Present != 4 {
		t.Errorf("Bad obj.Present: %v", obj.Present)
	}
	if obj.Absent != 2 {
		t.Errorf("Bad obj.Absent: %v", obj.Absent)
	}
	if obj.Unknown != 1 {
		t.Errorf("Bad obj.Unknown: %v", obj.Unknown)
	}
	if obj.ParticipantsName != "電三太郎,電四郎,電六郎,電七郎" {
		t.Errorf("Bad obj.ParticipantsName: %v", obj.ParticipantsName)
	}
}

/*
 * 翌年扱いになる日程のテスト
 */
func TestParseCsv_normally_nextyear(t *testing.T) {
	var (
		objDay time.Time
		obj    *schedule
		exist  bool
		err    error
	)
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/chouseisan.csv")
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
		t.Errorf("Entry not found. date is %v", objDay)
		return
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
	if obj.ParticipantsName != "電一,電次郎,電四郎,電六郎,電七郎" {
		t.Errorf("Bad obj.ParticipantsName: %v", obj.ParticipantsName)
	}
}

/*
 * カラムの無いケース
 */
func TestParseCsv_normally_no_col(t *testing.T) {
	var (
		objDay time.Time
		obj    *schedule
		exist  bool
		err    error
	)
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/no_col.csv")
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
		t.Errorf("Entry not found. date is %v", objDay)
		return
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
}

/*
 * 行の無いケース
 */
func TestParseCsv_normally_no_row(t *testing.T) {
	var (
		objDay time.Time
		obj    *schedule
		exist  bool
		err    error
	)
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/no_row.csv")
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
		return
	}
}

/*
 * 日程カラムのフォーマット不正データ
 */
func TestParseCsv_invalid_dateformat(t *testing.T) {
	var (
		// objDay time.Time
		// obj    *schedule
		// exist  bool
		err error
	)
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	//テストデータはファイルから読む
	testdata, err := os.Open("testdata/invalid.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer testdata.Close()

	tz, _ := time.LoadLocation("Asia/Tokyo")
	today := time.Date(2016, time.December, 1, 0, 0, 0, 0, tz)
	parseCsv(c, testdata, today)
	//パースできていればok
}