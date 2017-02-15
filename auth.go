package main

import (
	"database/sql"
	//_ "github.com/go-sql-driver/mysql"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
)

var pass_key string = "kdfaoajdfa"

func userAuth(repo_id int, user, pass string) int {
	db := db_open()
	defer db.Close()

	rel_pass := HashPass(pass)

	//查找用户是否存在
	ssql := "select id from users where user=? and pass=?"

	var uid int
	err := db.QueryRow(ssql, user, rel_pass).Scan(&uid)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if uid < 1 {
		return 0
	}

	//查看用户权限
	if !db_perm_has(repo_id, uid, 1) {
		return 0
	}

	return uid
}

func adminAuth(user, pass string) int {
	db := db_open()
	defer db.Close()

	rel_pass := HashPass(pass)

	//查找用户是否存在
	ssql := "select id from users where user=? and pass=? and isadmin=1"

	var uid int
	err := db.QueryRow(ssql, user, rel_pass).Scan(&uid)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if uid < 1 {
		return 0
	}

	return uid
}

func HashPass(in string) string {
	mac := hmac.New(md5.New, []byte(pass_key))
	mac.Write([]byte(in))
	out := make([]byte, 0, 40)
	return hex.EncodeToString(mac.Sum(out))
}
