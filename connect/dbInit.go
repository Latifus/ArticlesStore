package connect

import "database/sql"

func DBConnection() *sql.DB {
	db, errr := sql.Open("mysql", "root:50151832l@tcp(127.0.0.1:3306)/project_db")
	if errr != nil {
		panic(errr)
	}
	defer db.Close()
	return db
}
