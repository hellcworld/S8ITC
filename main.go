package main

import (
	"code.google.com/p/mahonia"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type StockRecord struct {
	Stocknum string
	// 0 30 31 3 4 5 8 9
	Stockname string
	Date      string
	Time      string
	Timeprice string
	Turnover  string
}
type Records struct {
	gorm.Model
	Date string `gorm:"index;not null;unique"`
	//turnover
	Szzs  template.HTML
	Hs300 template.HTML
	Sz50  template.HTML
	Zz500 template.HTML
	Szcz  template.HTML
	Sz100 template.HTML
	Zxbz  template.HTML
	Cybz  template.HTML
}

func GetDataHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if len(r.Form) > 0 {
		stocknum := string(r.Form["stocknum"][0])
		fmt.Fprintln(w, getdata(stocknum))
	}
}
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	get8index()
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Println(err)
	}
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic("Connect database failed!")
	}
	defer db.Close()
	var rs []Records
	db.Order("date desc").Find(&rs)
	data := struct{ RS []Records }{RS: rs}
	err = tmpl.Execute(w, data)

	if err != nil {
		log.Println(err)
	}
}
func main() {
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic("Connect database failed!")
	}
	defer db.Close()
	db.AutoMigrate(&Records{})
	db.AutoMigrate(&StockRecord{})
	go func() {
		for {
			get8index()
			time.Sleep(time.Second * 22)
		}
	}()
	http.HandleFunc("/getdata", GetDataHandler)
	http.HandleFunc("/", IndexHandler)
	http.ListenAndServe("0.0.0.0:8080", nil)
}
func get8index() {
	sr1 := getdata("000001")
	sr2 := getdata("000300")
	sr3 := getdata("000016")
	sr4 := getdata("000905")
	sr5 := getdata("399001")
	sr6 := getdata("399004")
	sr7 := getdata("399005")
	sr8 := getdata("399006")
	tmp := Records{
		Date:  sr8.Date,
		Szzs:  template.HTML(sr1.Timeprice + "<br>" + sr1.Turnover),
		Hs300: template.HTML(sr2.Timeprice + "<br>" + sr2.Turnover),
		Sz50:  template.HTML(sr3.Timeprice + "<br>" + sr3.Turnover),
		Zz500: template.HTML(sr4.Timeprice + "<br>" + sr4.Turnover),
		Szcz:  template.HTML(sr5.Timeprice + "<br>" + sr5.Turnover),
		Sz100: template.HTML(sr6.Timeprice + "<br>" + sr6.Turnover),
		Zxbz:  template.HTML(sr7.Timeprice + "<br>" + sr7.Turnover),
		Cybz:  template.HTML(sr8.Timeprice + "<br>" + sr8.Turnover),
	}
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic("Connect database failed!")
	}
	defer db.Close()
	db.Where(Records{Date: sr8.Date}).Assign(tmp).FirstOrCreate(&tmp)
}
func getdata(stocknum string) StockRecord {
	var resp *http.Response
	var err error
	if strings.HasPrefix(stocknum, "3") {
		resp, err = http.Get("http://hq.sinajs.cn/list=sz" + stocknum)
	} else {
		resp, err = http.Get("http://hq.sinajs.cn/list=sh" + stocknum)
	}
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	buf, _ := ioutil.ReadAll(resp.Body)
	s1 := string(buf)
	dec := mahonia.NewDecoder("gbk")
	s1, _ = dec.ConvertStringOK(s1)
	slice01 := strings.Split(s1, "=\"")
	slice02 := strings.Split(slice01[1], "\";")
	slice03 := strings.Split(slice02[0], ",")

	tmpnum, err := strconv.ParseFloat(slice03[9], 32)
	if err != nil {
		log.Println(err)
	}
	tmpnum = tmpnum / 10000 / 10000
	slice03[9] = strconv.FormatFloat(tmpnum, 'f', 0, 32)
	slice03[9] += "亿"

	tmpnum, err = strconv.ParseFloat(slice03[3], 32)
	if err != nil {
		log.Println(err)
	}
	slice03[3] = strconv.FormatFloat(tmpnum, 'f', 0, 32)

	stock := StockRecord{
		Stocknum:  stocknum,
		Stockname: slice03[0],
		Date:      slice03[30],
		Time:      slice03[31],
		Timeprice: slice03[3],
		Turnover:  slice03[9],
	}

	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic("Connect database failed!")
	}
	defer db.Close()
	db.Where(StockRecord{Date: stock.Date, Stocknum: stock.Stocknum}).Assign(stock).FirstOrCreate(&stock)
	return stock
}
