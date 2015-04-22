package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"stockManagement/lib"
	"strings"
)

var templates = template.Must(template.ParseGlob("tmpl/*.html"))

func renderTemplate(w http.ResponseWriter, path string, scripts []sqlAdapter.Stock) {
	t, err := template.ParseFiles("tmpl/" + path + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, scripts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func LaterrenderTemplate(w http.ResponseWriter, path string, scripts []sqlAdapter.Stock) {
	err := templates.ExecuteTemplate(w, path+".html", scripts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	var query string
	if r.FormValue("type") == "addScript" {
		user := r.FormValue("user")
		price := r.FormValue("price")
		company := r.FormValue("company")
		quantity := r.FormValue("quantity")
		trade := r.FormValue("trade")
		nse := r.FormValue("nse")
		if !(len(user) == 0 || len(price) == 0 || len(company) == 0 || len(quantity) == 0 || len(nse) == 0) {
			query = fmt.Sprintf("insert into scripts (user, nse, bse, company, quantity, trade, date, price, current_price)  values (\"%s\", \"%s\", \"%s\", \"%s\", %s, %s, NOW(), %s, %s);", user, nse, "", company, quantity, trade, price, "1")
			sqlAdapter.ExecuteQuery(query)
		}
		query = "SELECT * FROM scripts"
		renderTemplate(w, "home", sqlAdapter.GetAllScripts(query))
	} else if len(r.FormValue("query")) != 0 {
		query := r.FormValue("query")
		if len(query) == 0 {
			query = "SELECT * FROM scripts"
		}
		if !(strings.ToUpper(strings.Fields(query)[0]) == "SELECT") {
			sqlAdapter.ExecuteQuery(query)
		}
		renderTemplate(w, "home", sqlAdapter.GetAllScripts(query))
	} else {
		query = "SELECT * FROM scripts"
		renderTemplate(w, "home", sqlAdapter.GetAllScripts(query))
	}
}
func attachHandlers() {
	http.HandleFunc("/", homeHandler)
}
func RunServer() {
	attachHandlers()
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	sqlAdapter.ConnectDB()
	RunServer()
	defer sqlAdapter.CloseDB()
}
