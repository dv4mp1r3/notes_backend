package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func OpenDB(path string) (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite3", path)
	return db, err
}

func queryStatement(query string, args ...any) (*sql.Rows, error) {
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt.Query(args...)
}

func execStatement(query string, args ...any) (bool, error) {

	stmt, err := db.Prepare(query)
	if err != nil {
		return false, err
	}
	_, err = stmt.Exec(args...)

	if err != nil {
		return false, err
	}
	return true, err
}

func IsAccountCorrect(user User) (int, error) {
	var id int
	rows, err := queryStatement("SELECT id FROM user WHERE username=? AND password=?", user.Username, user.Password)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&id)
	}
	return id, err
}

func GetUserResources(user User) []Resource {
	result := []Resource{}
	rows, err := queryStatement(
		"SELECT id,name,data,user_id FROM resources WHERE resources.user_id = ?",
		user.ID,
	)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var res Resource
		rows.Scan(&res.ID, &res.Name, &res.Data, &res.UserId)
		result = append(result, res)
	}
	return result
}

func InsertResource(res Resource) (bool, error) {
	return execStatement("INSERT INTO resources (name, data, user_id) VALUES (?,?,?)", res.Name, res.Data, res.UserId)
}

func UpdateResource(res Resource) (bool, error) {
	return execStatement("UPDATE resources SET name=?,data=? WHERE id=?", res.Name, res.Data, res.ID)
}

func DeleteResource(res Resource) (bool, error) {
	return execStatement("DELETE FROM resources WHERE id=?", res.ID)
}
