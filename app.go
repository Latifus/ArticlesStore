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
type Exception struct {
	Message string
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
		panic(err)
	}

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}

	defer db.Close()

	res, error := db.Query("SELECT * FROM `items` ORDER BY `creationDate` DESC;")
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

func getAddItemPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("front/add.html", "front/header.html", "front/footer.html")
	if err != nil {
		panic(err)
	}

	temp.ExecuteTemplate(w, "create", nil)
}

func saveItem(w http.ResponseWriter, r *http.Request) {

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
	return criteria != ""
}

func getRegisterPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("front/login.html", "front/footer.html", "front/headerForAuth.html",)
	if err != nil {
		panic(err)
	}

	temp.ExecuteTemplate(w, "register", nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}
	defer db.Close()

	if checkCriteria(username) && checkCriteria(email) && checkCriteria(password) {
		insert, error := db.Query(fmt.Sprintf("INSERT INTO `users` (`username`, `email`, `password`) VALUES('%s', '%s',  '%s')", username, email, password))
		if error != nil {
			panic(error)
		}
		defer insert.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("front/login.html", "front/footer.html", "front/headerForAuth.html",)
	if err != nil {
		panic(err)
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}
	defer db.Close()

	res, err := db.Query("SELECT `email`, `password` FROM `users`")
	if err != nil {
		panic(errr)
	}

	if checkCriteria(email) && checkCriteria(password) {

		found := false
		for res.Next() {
			var user User
			err = res.Scan(&user.Email, &user.Password)
			if err != nil {
				panic(err)
			}

			if user.Email == email && user.Password == password {
				found = true
				break
			}
		}

		if found {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}else {
		errorMessage := "username or password must not be empty!"
		ex := Exception{Message: errorMessage}
		temp.ExecuteTemplate(w, "login", ex)
	}
}

func main() {
	rtr := mux.NewRouter()
	http.Handle("/", rtr)

	rtr.HandleFunc("/", getHomePage).Methods("GET")
	rtr.HandleFunc("/add", getAddItemPage).Methods("GET")
	rtr.HandleFunc("/save_article", saveItem).Methods("POST")
	rtr.HandleFunc("/register", getRegisterPage).Methods("GET")
	rtr.HandleFunc("/save_user", register).Methods("POST")
	rtr.HandleFunc("/login", login).Methods("GET")

	http.ListenAndServe(":8181", nil)
}
