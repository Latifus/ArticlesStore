package main

import (
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func getHomePage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("front/home.html", "front/header.html", "front/footer.html",)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	temp.ExecuteTemplate(w, "index", nil)
}

func main() {

	rtr := mux.NewRouter()
	http.Handle("/", rtr)

	rtr.HandleFunc("/", getHomePage).Methods("GET")

	http.ListenAndServe(":8181", nil)
}
