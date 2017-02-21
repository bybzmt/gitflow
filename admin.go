package main

import (
	//"database/sql"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

func catch_error(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		if e, ok := err.(error); ok {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		trace := make([]byte, 1024)
		count := runtime.Stack(trace, true)
		m1 := fmt.Sprintf("Recover from panic: %s\n", err)
		m2 := fmt.Sprintf("Stack of %d bytes: %s\n", count, trace)

		log.Println(r.Method, r.URL.RequestURI(), 500, "[[", m1, m2, "]]")
	} else {
		log.Println(r.Method, r.URL.RequestURI(), 200)
	}
}

func page_admin_auth(w http.ResponseWriter, r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	if !ok {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"USER LOGIN\"")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return false
	}

	//用户id
	user_id := adminAuth(user, pass)
	if user_id < 1 {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"USER LOGIN\"")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return false
	}
	return true
}

func page_admin_index(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	repos := db_repos_getall()

	tmpl := load_tpl("index.tpl")

	var data = struct {
		Repos     []RepoRow
		ReposBase string
	}{
		Repos:     repos,
		ReposBase: "http://" + r.Host + r.RequestURI + "repos/",
	}

	tmpl.Execute(w, data)
}

func page_admin_users(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	users := db_users_getall()

	tmpl := load_tpl("users.tpl")

	var data = struct {
		Users []UserRow
	}{
		Users: users,
	}

	tmpl.Execute(w, data)
}

func page_admin_useradd(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	//TODO 注册开关判断

	tmpl := load_tpl("useradd.tpl")

	var data = struct {
	}{}

	tmpl.Execute(w, data)
}

func page_admin_useradd_do(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	//TODO 注册开关判断

	user := r.FormValue("user")
	pass := r.FormValue("pass")

	db := db_open()
	defer db.Close()

	rel_pass := HashPass(pass)

	ssql := "insert into users (`user`, `pass`, `isadmin`) value (?, ?, 0)"
	res, err := db.Exec(ssql, user, rel_pass)
	if err != nil {
		panic(err)
	}

	uid, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	//将第1个注册的用户置为管理员
	if uid == 1 {
		ssql = "update users set isadmin = 1 where id = ?"
		db.Exec(ssql, uid)
	}

	show_confirm(w, r, "添加成功", "/", "")
}

func page_admin_useredit(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	uid := r.FormValue("uid")
	rel_uid, _ := strconv.ParseInt(uid, 10, 32)

	user := db_user_get(int(rel_uid))
	if user == nil {
		show_confirm(w, r, "用户不存在", "/admin/users", "")
		return
	}

	repos := db_repos_getall()
	rules := ReposRules
	perms := db_perms_getall(int(rel_uid))

	rel_perms := make(map[int]map[int]bool)
	for _, perm := range perms {
		_, ok := rel_perms[perm.Rid]
		if !ok {
			rel_perms[perm.Rid] = make(map[int]bool)
		}

		rel_perms[perm.Rid][perm.Rule] = true
	}

	tmpl := load_tpl("useredit.tpl")

	var data = struct {
		User  UserRow
		Repos []RepoRow
		Rules []RuleRow
		Perms map[int]map[int]bool
	}{
		User:  *user,
		Repos: repos,
		Rules: rules,
		Perms: rel_perms,
	}

	tmpl.Execute(w, data)
}
func page_admin_useredit_do(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	uid := r.FormValue("uid")
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	isadmin := r.FormValue("isadmin")

	if uid == "1" && isadmin != "1" {
		show_confirm(w, r, "root必需是管理员", "/admin/users", "")
	}

	perms, _ := r.Form["perms[]"]

	db := db_open()
	defer db.Close()

	rel_isadmin, _ := strconv.ParseInt(isadmin, 10, 8)

	if pass != "" {
		rel_pass := HashPass(pass)

		ssql := "update users set `user` = ?, `pass` = ?, `isadmin` = ? where id = ?"
		_, err := db.Exec(ssql, user, rel_pass, rel_isadmin, uid)
		if err != nil {
			panic(err)
		}
	} else {
		ssql := "update users set `user` = ?, `isadmin` = ? where id = ?"
		_, err := db.Exec(ssql, user, rel_isadmin, uid)
		if err != nil {
			panic(err)
		}
	}

	//先删除
	ssql := "delete from perms where `uid` = ?"
	_, err := db.Exec(ssql, uid)
	if err != nil {
		panic(err)
	}

	//再添加
	ssql = "insert into perms (`rid`, `uid`, `rule`) value(?,?,?)"
	stmt, err := db.Prepare(ssql)
	if err != nil {
		panic(err)
	}

	for _, perm := range perms {
		t := strings.Split(perm, ":")
		if len(t) != 2 {
			w.Write([]byte("request bad"))
			return
		}
		rid, err := strconv.ParseInt(t[0], 10, 8)
		if err != nil {
			panic(err)
		}
		rule, err := strconv.ParseInt(t[1], 10, 8)
		if err != nil {
			panic(err)
		}

		//TODO 验证rule合法性

		_, err = stmt.Exec(rid, uid, rule)
		if err != nil {
			panic(err)
		}
	}

	show_confirm(w, r, "编辑成功", "/admin/users", "")
}

func page_admin_userdel(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	uid := r.FormValue("uid")

	show_confirm(w, r, "您确认删除吗?", "/admin/userdel_do?uid="+uid, "/admin/users")
}

func page_admin_userdel_do(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	uid := r.FormValue("uid")

	if uid == "1" {
		show_confirm(w, r, "root不能删除", "/admin/users", "")
		return
	}

	db := db_open()
	defer db.Close()

	ssql := "delete from perms where uid = ?"
	_, err := db.Exec(ssql, uid)
	if err != nil {
		panic(err)
	}

	ssql = "delete from users where id = ?"
	_, err = db.Exec(ssql, uid)
	if err != nil {
		panic(err)
	}

	show_confirm(w, r, "删除成功", "/admin/users", "")
}

func page_admin_repoadd(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	tmpl := load_tpl("repoadd.tpl")

	var data = struct {
	}{}

	tmpl.Execute(w, data)
}

func page_admin_repoadd_do(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	name := r.FormValue("name")
	about := r.FormValue("about")
	branch_names := r.FormValue("branch_names")
	tag_names := r.FormValue("tag_names")
	branch_locks := r.FormValue("branch_locks")

	db := db_open()
	defer db.Close()

	ssql := "insert into repositories (`name`, `message`) values (?, ?)"
	res, err := db.Exec(ssql, name, about)
	if err != nil {
		panic(err)
	}

	repo_id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	db_config_set(repo_id, "branch_names", branch_names)
	db_config_set(repo_id, "tag_names", tag_names)
	db_config_set(repo_id, "branch_locks", branch_locks)

	show_confirm(w, r, "添加成功", "/", "")
}

func page_admin_repoedit(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	uid := r.FormValue("rid")
	rel_uid, _ := strconv.ParseInt(uid, 10, 32)

	repo := db_repos_get(int(rel_uid))
	if repo == nil {
		show_confirm(w, r, "库不存在", "/", "")
		return
	}

	tmpl := load_tpl("repoedit.tpl")

	branch_names := db_config_get(repo.Id, "branch_names")
	tag_names := db_config_get(repo.Id, "tag_names")
	branch_locks := db_config_get(repo.Id, "branch_locks")

	var data = struct {
		Repo        RepoRow
		BranchNames string
		BranchLocks string
		TagNames    string
	}{
		Repo:        *repo,
		BranchNames: branch_names,
		BranchLocks: branch_locks,
		TagNames:    tag_names,
	}

	tmpl.Execute(w, data)
}

func page_admin_repoedit_do(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	rid := r.FormValue("rid")
	name := r.FormValue("name")
	about := r.FormValue("about")
	branch_names := r.FormValue("branch_names")
	tag_names := r.FormValue("tag_names")
	branch_locks := r.FormValue("branch_locks")

	repo_id, _ := strconv.ParseInt(rid, 10, 32)

	db := db_open()
	defer db.Close()

	ssql := "update repositories set `name` = ?, `message` = ? where id = ?"
	_, err := db.Exec(ssql, name, about, rid)
	if err != nil {
		panic(err)
	}

	db_config_set(repo_id, "branch_names", branch_names)
	db_config_set(repo_id, "tag_names", tag_names)
	db_config_set(repo_id, "branch_locks", branch_locks)

	show_confirm(w, r, "编辑成功", "/", "")
}

func page_admin_repodel(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	rid := r.FormValue("rid")

	show_confirm(w, r, "您确认删除吗?", "/admin/repodel_do?rid="+rid, "/admin/users")
}

func page_admin_repodel_do(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	if !page_admin_auth(w, r) {
		return
	}

	rid := r.FormValue("rid")

	db := db_open()
	defer db.Close()

	ssql := "delete from repositories where id = ?"
	_, err := db.Exec(ssql, rid)
	if err != nil {
		panic(err)
	}

	show_confirm(w, r, "删除成功", "/", "")
}
