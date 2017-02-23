package main

import (
	"database/sql"
	//_ "github.com/go-sql-driver/mysql"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"regexp/syntax"
	"strings"
)

//项目权限设置
const (
	//访问项目
	REPOS_RULE_READ = 1
	//修改项目
	REPOS_RULE_WRITE = 2
	//修改锁定分支
	REPOS_RULE_LOCK = 3
	//删除分支"
	REPOS_RULE_DEL = 4
	//新建tag
	REPOS_RULE_TAG_ADD = 5
	//修改/删除tag
	REPOS_RULE_TAG_EDIT = 6
)

var ReposRules = []RuleRow{
	{Id: REPOS_RULE_READ, About: "访问项目"},
	{Id: REPOS_RULE_WRITE, About: "修改项目"},
	{Id: REPOS_RULE_DEL, About: "删除分支"},
	{Id: REPOS_RULE_LOCK, About: "修改锁定分支"},
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

func MatchPartten(parttens, name string, empty bool) bool {
	parttens = strings.Trim(parttens, "\r\n\t ")
	if parttens == "" {
		return empty
	}

	tmps := strings.Split(parttens, "\n")
	for _, tmp := range tmps {
		//处理dot
		regx, err := syntax.Parse("^"+strings.Trim(tmp, "\r\n\t ")+"$", syntax.PerlX|syntax.MatchNL|syntax.UnicodeGroups)
		if err != nil {
			panic(err.Error())
		}

		reg, err := regexp.Compile(regx.String())
		if err != nil {
			panic(err.Error())
		}

		if reg.MatchString(name) {
			return true
		}
	}

	return false
}
