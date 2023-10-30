package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Item struct {
	ID           uint16
	AuthorName   string
	CreationDate string
	Likes        int64
	Title        string
	Description  string
}
var items = []Item{}

func formatDate(date string) string  {
	t, err := time.Parse("2006-01-02 15:04:05", date)
    if err != nil {
        return "Invalid Date"
    }
    return t.Format("2006 January 02")
}

func getHomePage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("front/home.html", "front/header.html", "front/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}

	defer db.Close()

	res, error := db.Query("SELECT * FROM `items`")
	if error != nil {
		panic(error)
	}

	items = []Item{}
	for res.Next() {
		var item Item
		error = res.Scan(&item.ID, &item.AuthorName, &item.CreationDate, &item.Likes, &item.Title, &item.Description)
		if error != nil {
			panic(error)
		}

		item.CreationDate = formatDate(item.CreationDate)

		items = append(items, item)
	}

	temp.ExecuteTemplate(w, "index", items)
}

func main() {
	rtr := mux.NewRouter()
	http.Handle("/", rtr)

	rtr.HandleFunc("/", getHomePage).Methods("GET")

	http.ListenAndServe(":8181", nil)
}
