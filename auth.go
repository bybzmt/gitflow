package main

import (
	"database/sql"
	//_ "github.com/go-sql-driver/mysql"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
)

//项目权限设置
const (
	//访问项目
	REPOS_RULE_READ = 1
	//修改项目
	REPOS_RULE_WRITE = 2
	//新建分支
	REPOS_RULE_ADD = 3
	//删除分支"
	REPOS_RULE_DEL = 4
	//修改锁定分支
	REPOS_RULE_LOCK = 5
	//强型推送
	REPOS_RULE_FORCE = 6
	//新建tag
	REPOS_RULE_TAG_ADD = 7
	//修改/删除tag
	REPOS_RULE_TAG_EDIT = 8
)

var ReposRules = []RuleRow{
	{Id: REPOS_RULE_READ, About: "访问项目"},
	{Id: REPOS_RULE_WRITE, About: "修改项目"},
	{Id: REPOS_RULE_ADD, About: "新建分支"},
	{Id: REPOS_RULE_DEL, About: "删除分支"},
	{Id: REPOS_RULE_LOCK, About: "修改锁定分支"},
	{Id: REPOS_RULE_FORCE, About: "强型推送"},
	{Id: REPOS_RULE_TAG_ADD, About: "新建tag"},
	{Id: REPOS_RULE_TAG_EDIT, About: "修改/删除tag"},
}

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
