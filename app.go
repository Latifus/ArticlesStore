package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Item struct {
	ID           uint16
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
	Token    string
}
type Tag struct {
	ID      uint16
	Tagname string
	Item_id uint16
}
type ItemWithUsername struct {
	ID           int
	CreationDate string
	Likes        int
	Title        string
	Description  string
	Username     string // New field to store the username
}
type Exception struct {
	Message string
}
type ItemWithTags struct {
	ID           int
	Likes        int
	Tagsname     string
	CreationDate string
	Titel        string
	Description  string
}

var secretKey string = "ZrStAfUuqTM6eTuhacT9JCfUQp9QkHnZ"

func (u *User) setToken(token string) {
	u.Token = token
}

func (u User) getToken() string {
	return u.Token
}

func generateToken(userID int) string {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	fmt.Println("saved id in claim")
	fmt.Println(claims["user_id"])

	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		panic(err)
	}

	return tokenString
}

func formatDate(date string) string {
	t, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		return "Invalid Date"
	}
	return t.Format("2006 January 02")
}

func checkToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid token signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("Invalid token")
}

func CheckAuthentication(tokenString string) bool {

	if tokenString == "" {
		return false
	}

	_, err := checkToken(tokenString)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func getHomePage(w http.ResponseWriter, r *http.Request) {
	var isAuthenticated bool
	cookie, err := r.Cookie("token")
	if err != nil {
		isAuthenticated = false
	} else {
		token := cookie.Value
		isAuthenticated = CheckAuthentication(token)
	}

	temp, err := template.ParseFiles("front/home.html", "front/header.html", "front/headerIfNonAuthorize.html", "front/footer.html")
	if err != nil {
		panic(err)
	}

	data := struct {
		IsAuthenticated bool
		Items           []ItemWithUsername
		Tags            []Tag
	}{
		IsAuthenticated: isAuthenticated,
	}

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}
	defer db.Close()

	query := `
        SELECT i.id, i.creationDate, i.likes, i.title, i.description, u.username
        FROM items i
        JOIN users u ON i.user_id = u.id
        ORDER BY i.creationDate DESC;
    `

	res, error := db.Query(query)
	if error != nil {
		panic(error)
	}

	for res.Next() {
		var item ItemWithUsername
		err = res.Scan(&item.ID, &item.CreationDate, &item.Likes, &item.Title, &item.Description, &item.Username)
		if err != nil {
			panic(err)
		}
		item.Description = cutDescription(item.Description)
		item.CreationDate = formatDate(item.CreationDate)

		data.Items = append(data.Items, item)
	}

	tag, fail := db.Query("SELECT * FROM `tags`;")
	if fail != nil {
		panic(fail)
	}

	data.Tags = []Tag{}
	for tag.Next() {
		var t Tag
		fail = tag.Scan(&t.ID, &t.Tagname, &t.Item_id)
		if fail != nil {
			panic(fail)
		}
		data.Tags = append(data.Tags, t)
	}

	temp.ExecuteTemplate(w, "index", data)
}

func updateLikes(w http.ResponseWriter, r *http.Request) {
	itemID := r.FormValue("itemID")
	likes := r.FormValue("likes")

	db, err := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, _ = db.Exec("UPDATE items SET likes = ? WHERE id = ?", likes, itemID)

	w.WriteHeader(http.StatusOK)
}

func cutDescription(desc string) string {
	words := strings.Fields(desc)
	if 10 > len(words) {
		return desc
	}
	return strings.Join(words[:10], " ") + "..."
}

func getAddItemPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("front/add.html", "front/header.html", "front/footer.html")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	tag, fail := db.Query("SELECT * FROM `tags`;")
	if fail != nil {
		panic(fail)
	}

	tags := []Tag{}
	for tag.Next() {
		var t Tag
		fail = tag.Scan(&t.ID, &t.Tagname, &t.Item_id)
		if fail != nil {
			panic(fail)
		}
		tags = append(tags, t)
	}

	temp.ExecuteTemplate(w, "create", tags)
}

func saveItem(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		panic(err)
	}

	userID := verifyToken(token.Value)

	title := r.FormValue("title")
	desc := r.FormValue("desc")
	date := time.Now().Format("2006-01-02 15:04:05")
	tags := r.Form["tags"]

	fmt.Print("addded tags: ")
	fmt.Print(tags)
	fmt.Println()
	fmt.Println(title + " : " + desc + " : " + date)

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}
	defer db.Close()

	insert, error := db.Exec(fmt.Sprintf("INSERT INTO `items` (`creationDate`, `likes`, `title`, `description`, `user_id`) VALUES('%s', '%d', '%s', '%s', '%d')", date, 0, title, desc, userID))
	if error != nil {
		panic(error)
	}

	itemID, err := insert.LastInsertId()
	if err != nil {
		panic(err)
	}

	for _, v := range tags {
		fmt.Println("elemtn: " + v)

		_, fail := db.Query(fmt.Sprintf("INSERT INTO `tags` (`tag_name`, `item_id`) VALUES('%s', '%d')", v, itemID))
		if fail != nil {
			panic(fail)
		}

	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkCriteria(criteria string) bool {
	return criteria != ""
}

func getRegisterPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("front/register.html", "front/footer.html", "front/headerForAuth.html")
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
	temp, err := template.ParseFiles("front/login.html", "front/footer.html", "front/headerForAuth.html")
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

	res, err := db.Query("SELECT `id`, `email`, `password` FROM `users`")
	if err != nil {
		panic(errr)
	}

	if checkCriteria(email) && checkCriteria(password) {

		found := false
		for res.Next() {
			var user User
			err = res.Scan(&user.ID, &user.Email, &user.Password)
			if err != nil {
				panic(err)
			}

			if user.Email == email && user.Password == password {
				found = true
				fmt.Println("user Id to send " + fmt.Sprint(user.ID))
				user.setToken(generateToken(int(user.ID)))

				fmt.Println("token: " + user.getToken())
				cooke := http.Cookie{
					Name:     "token",
					Value:    user.getToken(),
					Expires:  time.Now().Add(24 * time.Hour),
					HttpOnly: true,
				}
				http.SetCookie(w, &cooke)
				break
			}
		}

		if found {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	} else {
		errorMessage := "username or password must not be empty!"
		ex := Exception{Message: errorMessage}
		temp.ExecuteTemplate(w, "login", ex)
	}
}

func verifyToken(token string) int {
	claims, _ := checkToken(token)
	fmt.Println("claims: ")
	fmt.Println(claims)

	fmt.Println("variable in claim id:")
	fmt.Print(claims["user_id"])
	fmt.Println("--------")

	userID, _ := claims["user_id"].(float64)
	fmt.Println("Id from claims: " + fmt.Sprint(userID))
	return int(userID)
}

func checkUserPage(w http.ResponseWriter, r *http.Request) {
	token, _ := r.Cookie("token")
	fmt.Println(token)
	userId := verifyToken(token.Value)
	fmt.Println("ID: " + fmt.Sprint(userId))

	temp, err := template.ParseFiles("front/user.html", "front/footer.html", "front/header.html")
	if err != nil {
		panic(err)
	}

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}
	defer db.Close()

	res, exce := db.Query(fmt.Sprintf("SELECT * FROM `users` WHERE `id` = %d", userId))
	if err != nil {
		panic(exce)
	}

	var user User
	for res.Next() {
		err = res.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
		if err != nil {
			panic(err)
		}
	}

	temp.ExecuteTemplate(w, "user", user)
}

func logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getItemPage(w http.ResponseWriter, r *http.Request) {
	var isAuthenticated bool
	cookie, err := r.Cookie("token")
	if err != nil {
		isAuthenticated = false
	} else {
		token := cookie.Value
		isAuthenticated = CheckAuthentication(token)
	}

	temp, err := template.ParseFiles("front/show.html", "front/header.html", "front/headerIfNonAuthorize.html", "front/footer.html")
	if err != nil {
		panic(err)
	}

	data := struct {
		IsAuthenticated bool
		Item            []ItemWithTags
	}{
		IsAuthenticated: isAuthenticated,
	}

	vars := mux.Vars(r)

	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}
	defer db.Close()

	tag, fail := db.Query(`SELECT i.creationDate, i.likes, i.title, i.description, t.tag_name 
							FROM items i
							INNER JOIN tags t ON i.id = t.item_id
							WHERE i.id = ?;`, vars["id"])

	if fail != nil {
		panic(fail)
	}

	data.Item = []ItemWithTags{}
	for tag.Next() {
		
		var t ItemWithTags

		fail = tag.Scan(&t.CreationDate, &t.Likes, &t.Titel, &t.Description, &t.Tagsname)
		if fail != nil {
			panic(fail)
		}
		fmt.Print("new item: ")
		fmt.Print(t)
		fmt.Println("")

		data.Item = append(data.Item, t)

	}

	temp.ExecuteTemplate(w, "show", data)
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
	rtr.HandleFunc("/user", checkUserPage).Methods("GET")
	rtr.HandleFunc("/update-likes", updateLikes).Methods("POST")
	rtr.HandleFunc("/logout", logout).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}", getItemPage).Methods("GET")

	http.ListenAndServe(":8181", nil)
}
