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

type GitflowConfig struct {
	Develop       string
	Master        string
	BugfixPrefix  string
	BugfixExpr    string
	FeaturePrefix string
	FeatureExpr   string
	HotfixPrefix  string
	HotfixExpr    string
	ReleasePrefix string
	ReleaseExpr   string
	UserIsadmin   bool
}

func db_get_gitflow_config(repo_id int) *GitflowConfig {
	sql := "select master,develop,prefix_bugfix,prefix_feature,prefix_hotfix,prefix_release,expr_bugfix,expr_feature,expr_hotfix,expr_release from gitflow where rid = ?"

	db := db_open()
	defer db.Close()

	c := new(GitflowConfig)

	err := db.QueryRow(sql, repo_id).Scan(&c.Master, &c.Develop, &c.BugfixPrefix, &c.FeaturePrefix, &c.HotfixPrefix, &c.ReleasePrefix, &c.BugfixExpr, &c.FeatureExpr, &c.HotfixExpr, &c.ReleaseExpr)
	if err != nil {
		panic(err)
	}

	return c
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

func db_rules_getall() []RuleRow {
	ssql := "select id,about from rules"
	db := db_open()
	defer db.Close()

	rows, err := db.Query(ssql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var data []RuleRow

	for rows.Next() {
		t := RuleRow{}
		err := rows.Scan(&t.Id, &t.About)
		if err != nil {
			panic(err)
		}

		data = append(data, t)
	}

	return data
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

func db_branch_get_status(repo_id int, name string) int {
	db := db_open()
	defer db.Close()

	var status int

	ssql := "select `status` from branchs where rid = ? and branch = ?"
	err := db.QueryRow(ssql, hooks_ctx.RepoId, name).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return status
}

func db_branch_update_commit(repo_id int, name, commit string) {
	db := db_open()
	defer db.Close()

	ssql := "update branchs set commit = ? where rid = ? and branch = ?"
	_, err := db.Exec(ssql, commit, repo_id, name)
	if err != nil {
		panic(err)
	}
}

func db_history_branchs(repo_id int, prefix string) []string {
	ssql := "select branch from branchs where rid = ? and branch like ?"
	db := db_open()
	defer db.Close()

	rows, err := db.Query(ssql, repo_id, prefix+"%")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var data []string

	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			panic(err)
		}

		data = append(data, name)
	}

	return data
}

func db_commit_log_add(uid, repo_id int, branch, oldrev, newrev string) {
	db := db_open()
	defer db.Close()

	ssql := "INSERT INTO commit_log(rid, uid, branch, oldrev, newrev) VALUES(?,?,?,?)"
	_, err := db.Exec(ssql, repo_id, uid, branch, oldrev, newrev)
	if err != nil {
		panic(err)
	}
}
