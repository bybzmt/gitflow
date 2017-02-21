package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type RepoRow struct {
	Id      int
	Name    string
	Message string
}

type UserRow struct {
	Id      int
	User    string
	Isadmin int
}

type RuleRow struct {
	Id    int
	About string
}

type PermRow struct {
	Id   int
	Rid  int
	Uid  int
	Rule int
}

func db_open() *sql.DB {
	db, err := sql.Open("mysql", *dbdsn)
	if err != nil {
		panic(err)
	}

	return db
}

func db_config_get(repo_id int, key string) string {
	db := db_open()
	defer db.Close()

	ssql := "select `value` from config where rid = ? and `key`= ?"

	var val string

	err := db.QueryRow(ssql, repo_id, key).Scan(&val)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return val
}

func db_config_set(repo_id int64, key string, value string) {
	db := db_open()
	defer db.Close()

	ssql := "insert into config (`rid`, `key`, `value`) values(?, ?, ?) ON DUPLICATE KEY UPDATE `value` = ?"
	_, err := db.Exec(ssql, repo_id, key, value, value)
	if err != nil {
		panic(err)
	}
}

func db_find_repo_id(name string) int {
	db := db_open()
	defer db.Close()

	//查找用户是否存在
	ssql := "select id from repositories where name=?"

	var id int
	err := db.QueryRow(ssql, name).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return id
}

func db_repos_get(rid int) *RepoRow {
	ssql := "select id,name,message from repositories where id = ?"
	db := db_open()
	defer db.Close()

	t := RepoRow{}

	err := db.QueryRow(ssql, rid).Scan(&t.Id, &t.Name, &t.Message)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		} else {
			panic(err)
		}
	}

	return &t
}

func db_repos_getall() []RepoRow {
	ssql := "select id,name,message from repositories"
	db := db_open()
	defer db.Close()

	rows, err := db.Query(ssql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var data []RepoRow

	for rows.Next() {
		t := RepoRow{}
		err := rows.Scan(&t.Id, &t.Name, &t.Message)
		if err != nil {
			panic(err)
		}

		data = append(data, t)
	}

	return data
}

func db_users_getall() []UserRow {
	ssql := "select id,user,isadmin from users"
	db := db_open()
	defer db.Close()

	rows, err := db.Query(ssql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var data []UserRow

	for rows.Next() {
		t := UserRow{}
		err := rows.Scan(&t.Id, &t.User, &t.Isadmin)
		if err != nil {
			panic(err)
		}

		data = append(data, t)
	}

	return data
}

func db_user_get(uid int) *UserRow {
	ssql := "select id,user,isadmin from users where id = ?"
	db := db_open()
	defer db.Close()

	t := new(UserRow)

	err := db.QueryRow(ssql, uid).Scan(&t.Id, &t.User, &t.Isadmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}

	return t
}

func db_perms_getall(user_id int) []PermRow {
	ssql := "select id,rid,uid,rule from perms where uid = ?"
	db := db_open()
	defer db.Close()

	rows, err := db.Query(ssql, user_id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var data []PermRow

	for rows.Next() {
		t := PermRow{}
		err := rows.Scan(&t.Id, &t.Rid, &t.Uid, &t.Rule)
		if err != nil {
			panic(err)
		}

		data = append(data, t)
	}

	return data
}

func db_perm_has(repo_id, user_id, rule int) bool {
	db := db_open()
	defer db.Close()

	if user_id == 1 {
		return true
	}

	ssql := "select 1 from perms where rid = ? and uid = ? and rule = ?"
	has := 0
	err := db.QueryRow(ssql, repo_id, user_id, rule).Scan(&has)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			panic(err)
		}
	}

	return true
}

func db_commit_log_add(uid, repo_id int, branch, oldrev, newrev string) {
	db := db_open()
	defer db.Close()

	ssql := "INSERT INTO commit_log(rid, uid, branch, oldrev, newrev) VALUES(?,?,?,?,?)"
	_, err := db.Exec(ssql, repo_id, uid, branch, oldrev, newrev)
	if err != nil {
		panic(err)
	}
}
