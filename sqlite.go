package main

import (
	"database/sql"
	"log"

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

func GetUserResources(user User) []CategoryOutput {
	var result = []CategoryOutput{}
	rows, err := queryStatement(
		`SELECT r.id,r.name,r.data,r.user_id,r.category_id,r.icon,c.user_id, c.id,
			c.name as category_name, c.icon category_icon FROM resources r
			LEFT JOIN categories c on c.id = r.category_id
			WHERE r.user_id = ?`,
		user.ID,
	)
	if err != nil {
		return result
	}
	defer rows.Close()
	var tmp = map[int]*CategoryOutput{}
	for rows.Next() {
		var res Resource
		var cat CategoryOutput
		rows.Scan(
			&res.ID,
			&res.Name,
			&res.Data,
			&res.UserId,
			&res.CategoryId,
			&res.Icon,
			&cat.Category.UserId,
			&cat.Category.ID,
			&cat.Category.Name,
			&cat.Category.Icon,
		)
		if _, ok := tmp[res.CategoryId]; !ok {
			tmp[res.CategoryId] = &CategoryOutput{
				Category: Category{
					Name:   cat.Name,
					Icon:   cat.Icon,
					UserId: cat.UserId,
					ID:     cat.ID,
				},
				Resources: []Resource{},
			}
		}
		tmp[res.CategoryId].Resources = append(tmp[res.CategoryId].Resources, res)

	}
	for _, value := range tmp {
		result = append(result, *value)
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

func DeleteUser(username string) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("DeleteUser on begin transaction error: %s", err)
		return false, err
	}
	delResResult, err := statementResultAsBool(execStatement("DELETE FROM resources WHERE id=(select id from user where username = ?)", username))
	if err != nil {
		log.Fatalf("DeleteUser on delete resources error: %s", err)
		tx.Rollback()
		return delResResult, err
	}
	delUserResult, err := statementResultAsBool(execStatement("DELETE FROM user WHERE username=?", username))
	if err != nil {
		log.Fatalf("DeleteUser on delete user error: %s", err)
		tx.Rollback()
		return delUserResult, err
	}
	tx.Commit()
	return delResResult && delUserResult, nil
}

func InsertResource(res Resource) (int64, error) {
	success, err := execStatement("INSERT INTO resources (name, data, user_id, category_id, icon) VALUES (?,?,?,?,?)", res.Name, res.Data, res.UserId, res.CategoryId, res.Icon)
	if err != nil {
		return 0, err
	}
	return success.LastInsertId()
}

func InsertCategory(cat Category) (int64, error) {
	success, err := execStatement("INSERT INTO categories (name, user_id, icon) VALUES (?,?,?,?)", cat.Name, cat.UserId, cat.Icon)
	if err != nil {
		return 0, err
	}
	return success.LastInsertId()
}

func UpdateResource(res Resource) (bool, error) {
	return statementResultAsBool(execStatement("UPDATE resources SET name=?,data=? icon=? WHERE id=?", res.Name, res.Data, res.Icon, res.ID))
}

func UpdateCategory(cat Category) (bool, error) {
	return statementResultAsBool(execStatement("UPDATE categories SET name=?,icon=? WHERE id=?", cat.Name, cat.Icon, cat.ID))
}

func DeleteResource(id int) (bool, error) {
	return statementResultAsBool(execStatement("DELETE FROM resources WHERE id=?", id))
}

func DeleteCategory(id int) (bool, error) {
	return statementResultAsBool(execStatement("DELETE FROM categories WHERE id=?", id))
}
