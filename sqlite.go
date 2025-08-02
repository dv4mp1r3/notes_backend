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

func execStatement(query string, args ...any) (sql.Result, error) {

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt.Exec(args...)
}

func statementResultAsBool(result sql.Result, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	return true, err
}

func IsAccountCorrect(user User, salt string) (int, error) {
	var id int
	hash := HashPassword(user.Password, salt)
	rows, err := queryStatement("SELECT id FROM user WHERE username=? AND password=?", user.Username, hash)
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
		"SELECT id,name,data,user_id,icon FROM resources WHERE resources.user_id = ?",
		user.ID,
	)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var res Resource
		rows.Scan(&res.ID, &res.Name, &res.Data, &res.UserId, &res.Icon)
		result = append(result, res)
	}
	return result
}

func AddUser(username string, password string, salt string) (int64, error) {
	success, err := execStatement(
		"INSERT INTO user (username, password) VALUES (?,?)",
		username,
		HashPassword(password, salt),
	)
	if err != nil {
		return 0, err
	}
	return success.LastInsertId()
}

func InsertResource(res Resource) (int64, error) {
	success, err := execStatement("INSERT INTO resources (name, data, user_id, icon) VALUES (?,?,?,?)", res.Name, res.Data, res.UserId, res.Icon)
	if err != nil {
		return 0, err
	}
	return success.LastInsertId()
}

func UpdateResource(res Resource) (bool, error) {
	return statementResultAsBool(execStatement("UPDATE resources SET name=?,data=?, icon=? WHERE id=?", res.Name, res.Data, res.Icon, res.ID))
}

func DeleteResource(id int) (bool, error) {
	return statementResultAsBool(execStatement("DELETE FROM resources WHERE id=?", id))
}
