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
	Sell = iota
	Buy
	Total
	SumTotal
)

type Stock struct {
	NSEName           string    `nse`
	BSEName           string    `bse`
	Company           string    `company`
	Quantity          uint32    `quantity`
	Trade             int8      `trade`
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
func getCurrentPrice(symbol string) float32 {
	var temp float64
	doc, err := goquery.NewDocument("https://in.finance.yahoo.com/q?s=" + symbol + ".BO")
	if err != nil {
		log.Fatal(err)
	}
	temp, err = strconv.ParseFloat(doc.Find(".time_rtq_ticker").Text(), 32)
	return float32(temp)
}
func sumUpByCompany(scripts []Stock) []Stock {
	var sum_total Stock
	var total_current Stock
	total := []Stock{}

	sum_total.Trade = SumTotal
	sum_total.User = "Sum Total"
	total_current.NSEName = ""
	total_current.Trade = Total

	for _, value := range scripts {
		if len(total) != 0 && value.NSEName != total[len(total)-1].NSEName {
			total = append(total, total_current)
			total_current.Quantity = 0
			total_current.TotalValue = 0.0
			total_current.CurrentTotalValue = 0.0
		}
		total = append(total, value)
		total_current.User = "total"
		total_current.NSEName = value.NSEName
		total_current.BSEName = value.BSEName
		total_current.Company = value.Company
		total_current.CurrentPrice = value.CurrentPrice
		if value.Trade == Buy {
			total_current.Quantity = total_current.Quantity + value.Quantity
		} else if value.Trade == Sell {
			total_current.Quantity = total_current.Quantity - value.Quantity
		}
		total_current.TotalValue += value.Price * float32(value.Quantity)
		total_current.CurrentTotalValue += value.CurrentPrice * float32(value.Quantity)
		total_current.Difference = total_current.CurrentTotalValue - total_current.TotalValue
		total_current.Price = total_current.TotalValue / float32(total_current.Quantity)
		sum_total.TotalValue += value.Price * float32(value.Quantity)
		sum_total.CurrentTotalValue += value.CurrentPrice * float32(value.Quantity)
	}
	sum_total.Difference = sum_total.CurrentTotalValue - sum_total.TotalValue
	total = append(total, total_current)
	total = append(total, sum_total)
	return total
}
func GetCurrentPriceAll() {
	var script_name string
	rows, err := db.Query("select distinct nse from scripts;")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		err := rows.Scan(&script_name)
		price := getCurrentPrice(script_name)
		update_current_price_query := fmt.Sprintf("update scripts set current_price=%.02f where nse=\"%s\";", price, script_name)
		ExecuteQuery(update_current_price_query)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal("Error getting rows for allScripts")
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	rows.Close()
	defer rows.Close()
}
func getSlice(rows *sql.Rows, total bool) []Stock {
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
		//GetCurrentPrice(&current_script.CurrentPrice, current_script.NSEName)

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
	if total {
		return sumUpByCompany(scripts)
	}
	return scripts
}
func GetAllScripts(query string, total bool) []Stock {
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Error getting rows for allScripts")
	}
	allScripts := getSlice(rows, total)
	return allScripts
}
func ExecuteQuery(query string) {
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
