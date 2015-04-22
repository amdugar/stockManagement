package sqlAdapter

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"time"
)

const (
	All = iota
	Buy
	Sell
)

type Stock struct {
	NSEName           string    `nse`
	BSEName           string    `bse`
	Company           string    `company`
	Quantity          uint32    `quantity`
	Trade             uint8     `trade`
	Date              time.Time `date`
	Price             float32   `price`
	DisplayDate       string    `display`
	TotalValue        float32   `total`
	CurrentPrice      float32   `current`
	CurrentTotalValue float32   `currentvalue`
	Difference        float32   `diff`
	User              string    `user`
}

var db *sql.DB

func ConnectDB() {
	var err error
	db, err = sql.Open("mysql", "root:@/stocks?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to Database " + err.Error())
	}
}
func CloseDB() {
	db.Close()
}
func GetCurrentPrice(price *float32, symbol string) {
	var temp float64
	doc, err := goquery.NewDocument("https://in.finance.yahoo.com/q?s=" + symbol + ".BO")
	if err != nil {
		log.Fatal(err)
	}
	temp, err = strconv.ParseFloat(doc.Find(".time_rtq_ticker").Text(), 32)
	*price = float32(temp)
}
func getSlice(rows *sql.Rows) []Stock {
	scripts := []Stock{}
	var current_script Stock
	for rows.Next() {
		err := rows.Scan(
			&current_script.NSEName,
			&current_script.BSEName,
			&current_script.Company,
			&current_script.Quantity,
			&current_script.Trade,
			&current_script.Date,
			&current_script.Price,
			&current_script.CurrentPrice,
			&current_script.User)
		if err != nil {
			log.Fatal(err)
		}
		tempTime := current_script.Date
		GetCurrentPrice(&current_script.CurrentPrice, current_script.NSEName)

		current_script.DisplayDate = fmt.Sprintf("%d/%d/%d", tempTime.Day(), tempTime.Month(), tempTime.Year())
		current_script.TotalValue = current_script.Price * float32(current_script.Quantity)
		current_script.CurrentTotalValue = current_script.CurrentPrice * float32(current_script.Quantity)
		current_script.Difference = current_script.CurrentTotalValue - current_script.TotalValue
		scripts = append(scripts, current_script)
	}
	err := rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	rows.Close()
	defer rows.Close()
	return scripts
}
func GetAllScripts(query string) []Stock {
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Error getting rows for allScripts")
	}
	allScripts := getSlice(rows)
	return allScripts
}
func ExecuteQuery(query string) {
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
