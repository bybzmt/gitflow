package main

import (
	"database/sql"
	//_ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"regexp/syntax"
	"runtime"
)

type Result struct {
	Code int
	Msg  string
	Data *json.RawMessage
}

type BranchInfo struct {
	Name   string
	Info   string
	Status int
}

func catch_api_error(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		res := Result{}
		res.Code = 1
		res.Msg = fmt.Sprint(err)
		json.NewEncoder(w).Encode(res)

		trace := make([]byte, 1024)
		count := runtime.Stack(trace, true)
		m1 := fmt.Sprintf("Recover from panic: %s\n", err)
		m2 := fmt.Sprintf("Stack of %d bytes: %s\n", count, trace)

		log.Println(r.Method, r.URL.RequestURI(), 500, "[[", m1, m2, "]]")
	}
}

//得到本地所有分支
func page_branch_getall(w http.ResponseWriter, r *http.Request, ctx *Context) {
	defer catch_api_error(w, r)

	out, err := git_exec(ctx.RepoDir, "rev-parse", "--symbolic", "--branches")
	if err != nil {
		panic(errors.New(string(out)))
	}

	var b []string

	tmps := splitLines(string(out))
	for _, tmp := range tmps {
		b = append(b, tmp)
	}

	d, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	s := json.RawMessage(d)
	res := Result{Data: &s}
	json.NewEncoder(w).Encode(res)
}

//得到本地所有分支
func page_branch_has(w http.ResponseWriter, r *http.Request, ctx *Context) {
	defer catch_api_error(w, r)

	name := r.FormValue("name")

	out, err := git_exec(ctx.RepoDir, "rev-parse", "--symbolic", "--branches")
	if err != nil {
		panic(errors.New(string(out)))
	}

	find := false

	tmps := splitLines(string(out))
	for _, tmp := range tmps {
		if tmp == name {
			find = true
		}
	}

	d, err := json.Marshal(find)
	if err != nil {
		panic(err)
	}

	s := json.RawMessage(d)
	res := Result{Data: &s}
	json.NewEncoder(w).Encode(res)
}

//得到指定分支信息
func page_branch_getinfo(w http.ResponseWriter, r *http.Request, ctx *Context) {
	defer catch_api_error(w, r)

	name := r.FormValue("name")

	db := db_open()
	defer db.Close()

	var message string
	var status int

	ssql := "select `message`, `status` from branchs where rid = ? and branch = ?"

	err := db.QueryRow(ssql, ctx.RepoId, name).Scan(&message, &status)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	info := BranchInfo{}
	info.Name = name
	info.Info = message
	info.Status = status

	d, err := json.Marshal(info)
	if err != nil {
		panic(err)
	}

	s := json.RawMessage(d)
	res := Result{Data: &s}
	json.NewEncoder(w).Encode(res)
}

//新建分支
func page_branch_add(w http.ResponseWriter, r *http.Request, ctx *Context) {
	defer catch_api_error(w, r)

	name := r.FormValue("name")
	message := r.FormValue("message")

	res := Result{}

	conf := db_get_gitflow_config(ctx.RepoId)

	bugfix, err := syntax.Parse(conf.BugfixExpr, syntax.Perl)
	if err != nil {
		panic(err)
	}
	feature, err := syntax.Parse(conf.FeatureExpr, syntax.Perl)
	if err != nil {
		panic(err)
	}
	hotfix, err := syntax.Parse(conf.HotfixExpr, syntax.Perl)
	if err != nil {
		panic(err)
	}
	release, err := syntax.Parse(conf.ReleaseExpr, syntax.Perl)
	if err != nil {
		panic(err)
	}

	//找到基础分支
	var baseRef = ""

	//查看分支命名是否合法
	if regexp.MustCompile(bugfix.String()).MatchString(name) {
		baseRef = conf.Develop
	} else if regexp.MustCompile(feature.String()).MatchString(name) {
		baseRef = conf.Develop
	} else if regexp.MustCompile(hotfix.String()).MatchString(name) {
		baseRef = conf.Master
	} else if regexp.MustCompile(release.String()).MatchString(name) {
		baseRef = conf.Master
	} else {
		res.Code = 1
		res.Msg = "分支名不合法"
		json.NewEncoder(w).Encode(res)
		return
	}

	db := db_open()
	defer db.Close()

	//查看历史上有没有这个分支
	ssql := "select 1 from branchs where rid = ? and branch = ?"
	has := 0
	err = db.QueryRow(ssql, ctx.RepoId, name).Scan(&has)
	if err != nil {
		if err != sql.ErrNoRows {
			panic(err)
		}
	}

	if has > 0 {
		res.Code = 1
		res.Msg = "分支己存在"
		json.NewEncoder(w).Encode(res)
		return
	}

	//添加分支信息
	ssql = "insert into branchs (`rid`, `branch`, `message`, `status`) values(?, ?, ?, ?)"
	_, err = db.Exec(ssql, ctx.RepoId, name, message, 2)
	if err != nil {
		res.Code = 1
		res.Msg = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	//创建分支
	out, err := git_exec(ctx.RepoDir, "branch", name, baseRef)
	if err != nil {
		res.Code = 1
		res.Msg = string(out)
		json.NewEncoder(w).Encode(res)
		return
	}

	json.NewEncoder(w).Encode(res)
}

//分支删除
func page_branch_delete(w http.ResponseWriter, r *http.Request, ctx *Context) {
	defer catch_api_error(w, r)

	name := r.FormValue("name")

	res := Result{}

	//判断分支是否可删除 master develop不可删
	conf := db_get_gitflow_config(ctx.RepoId)

	if name == conf.Develop || name == conf.Master {
		res.Code = 1
		res.Msg = "分支" + name + "不可删除"
		json.NewEncoder(w).Encode(res)
		return
	}

	db := db_open()
	defer db.Close()

	//标记数据库中删除
	sql := "update branchs set `status` = 5 where rid = ? and branch = ? and `status` = 2"
	_, err := db.Exec(sql, ctx.RepoId, name)
	if err != nil {
		panic(err)
	}

	//删除分支
	out, err := git_exec(ctx.RepoDir, "branch", "-D", name)
	if err != nil {
		res.Code = 1
		res.Msg = string(out)
		json.NewEncoder(w).Encode(res)
		return
	}

	json.NewEncoder(w).Encode(res)
}

func page_branch_merged(w http.ResponseWriter, r *http.Request, ctx *Context) {
	defer catch_api_error(w, r)

	name := r.FormValue("name")

	res := Result{}

	status := 4

	db := db_open()
	defer db.Close()

	sql := "update branchs set status=? where rid = ? and branch = ? and `status` = 3"
	rs, err := db.Exec(sql, status, ctx.RepoId, name)
	if err != nil {
		panic(err)
	}

	aff, err := rs.RowsAffected()
	if err != nil {
		panic(err)
	}

	if aff < 1 {
		res.Code = 1
		res.Msg = "状态修改失败"
		json.NewEncoder(w).Encode(res)
		return
	}

	//删除分支
	out, err := git_exec(ctx.RepoDir, "branch", "-D", name)
	if err != nil {
		panic(errors.New(string(out)))
	}

	json.NewEncoder(w).Encode(res)
}

func page_branch_status(w http.ResponseWriter, r *http.Request, ctx *Context) {
	defer catch_api_error(w, r)

	name := r.FormValue("name")
	act := r.FormValue("act")

	res := Result{}

	var status int
	if act == "cancel" {
		status = 2
	} else if act == "apply" {
		status = 3
	} else if act == "pass" {
		status = 4
	} else {
		res.Code = 1
		res.Msg = "act 非法"
		json.NewEncoder(w).Encode(res)
		return
	}

	db := db_open()
	defer db.Close()

	sql := "update branchs set status=? where rid = ? and branch = ? and `status` in (2,3,4)"
	_, err := db.Exec(sql, status, ctx.RepoId, name)
	if err != nil {
		panic(err)
	}

	res.Code = 0
	json.NewEncoder(w).Encode(res)
}

//得到配置文件
func page_get_config(w http.ResponseWriter, r *http.Request, ctx *Context) {
	defer catch_api_error(w, r)

	conf := db_get_gitflow_config(ctx.RepoId)

	conf.UserIsadmin = db_perm_has(ctx.RepoId, ctx.UserId, 2)

	d, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}

	s := json.RawMessage(d)
	res := Result{Data: &s}
	json.NewEncoder(w).Encode(res)
}

//得到配置文件
func page_history_branchs(w http.ResponseWriter, r *http.Request, ctx *Context) {
	defer catch_api_error(w, r)

	prefix := r.FormValue("prefix")

	outs := db_history_branchs(ctx.RepoId, prefix)

	d, err := json.Marshal(outs)
	if err != nil {
		panic(err)
	}

	s := json.RawMessage(d)
	res := Result{Data: &s}
	json.NewEncoder(w).Encode(res)
}
