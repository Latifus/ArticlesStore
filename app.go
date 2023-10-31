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

type User struct {
	ID       uint16
	Username string
	Email    string
	Password string
}

var items = []Item{}

func formatDate(date string) string {
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

func addItem(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("front/add.html", "front/header.html", "front/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	temp.ExecuteTemplate(w, "create", nil)
}

func save(w http.ResponseWriter, r *http.Request) {

	author := r.FormValue("author")
	title := r.FormValue("title")
	desc := r.FormValue("desc")
	date := time.Now().Format("2006-01-02 15:04:05")

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}
	defer db.Close()

	insert, error := db.Query(fmt.Sprintf("INSERT INTO `items` (`authorName`, `creationDate`, `likes`, `title`, `description`) VALUES('%s', '%s', '%d', '%s', '%s')", author, date, 0, title, desc))
	if error != nil {
		panic(error)
	}
	defer insert.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkCriteria(criteria string) bool {
	return criteria != "";
}

func register(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	temp, err := template.ParseFiles("front/register.html", "front/footer.html")
	if err != nil {
		panic(err)
	}

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}
	defer db.Close()

	temp.ExecuteTemplate(w, "register", nil)

	if checkCriteria(username) && checkCriteria(email) && checkCriteria(password) {
		insert, error := db.Query(fmt.Sprintf("INSERT INTO `users` (`username`, `email`, `password`) VALUES('%s', '%s',  '%s')", username, email, password))
		if error != nil {
			panic(error)
		}
		defer insert.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}	

func main() {
	rtr := mux.NewRouter()
	http.Handle("/", rtr)

	rtr.HandleFunc("/", getHomePage).Methods("GET")
	rtr.HandleFunc("/add", addItem).Methods("GET")
	rtr.HandleFunc("/save_article", save).Methods("POST")
	rtr.HandleFunc("/register", register).Methods("GET")

	http.ListenAndServe(":8181", nil)
}
