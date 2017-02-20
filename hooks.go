package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var hooks_lock sync.Mutex
var sid = ""
var hook_update_url = ""
var hooks_ctx *Context

func hooks_start(ctx *Context) {
	hooks_lock.Lock()
	tmp := rand.Int63()
	sid = strconv.FormatInt(tmp, 16)
	hooks_ctx = ctx
	hook_update_url = ctx.BaseUrl + "/__hooks__/update"
}

func hooks_end() {
	hooks_lock.Unlock()
	sid = ""
	hooks_ctx = nil
}

func page_hooks_update(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	_sid := r.FormValue("sid")
	refname := r.FormValue("refname")
	oldrev := r.FormValue("oldrev")
	newrev := r.FormValue("newrev")
	newrev_type := r.FormValue("newrev_type")

	log.Println(_sid, oldrev, newrev, refname, newrev_type)

	if sid != _sid {
		w.Write([]byte("*** sid错误"))
		return
	}

	db := db_open()
	defer db.Close()

	if strings.HasPrefix(refname, "refs/heads/") {
		name := refname[len("refs/heads/"):]

		//判断是否是锁定分支
		branch_locks := db_config_get(hooks_ctx.RepoId, "branch_locks")
		if MatchPartten(branch_locks, name, false) {
			//锁定分支权限检查
			if !db_perm_has(hooks_ctx.RepoId, hooks_ctx.UserId, REPOS_RULE_LOCK) {
				w.Write([]byte("*** 您没有权限添加新分支！"))
				return
			}
		} else {
			//一般分支权限检查
			if !db_perm_has(hooks_ctx.RepoId, hooks_ctx.UserId, REPOS_RULE_WRITE) {
				w.Write([]byte("*** 您没有权限修改！"))
				return
			}
		}

		if newrev_type == "commit" {
			//TODO 强型推送检查

			//检查分支命名是否合法
			branch_names := db_config_get(hooks_ctx.RepoId, "branch_names")
			if !MatchPartten(branch_names, name, true) {
				w.Write([]byte("*** 分支命名不合法！"))
				return
			}
		} else if newrev_type == "delete" {
			//tag修改权限检查
			if !db_perm_has(hooks_ctx.RepoId, hooks_ctx.UserId, REPOS_RULE_DEL) {
				w.Write([]byte("*** 您没有权限删除分支！"))
				return
			}
		} else {
			w.Write([]byte("*** newrev_type 错误！"))
			return
		}

		//提交日志
		db_commit_log_add(hooks_ctx.UserId, hooks_ctx.RepoId, refname, oldrev, newrev)

	} else if strings.HasPrefix(refname, "refs/tags/") {
		name := refname[len("refs/tags/"):]

		if newrev_type == "commit" || newrev_type == "tag" {
			//commit == un-annotated tag (没有备注的tag)
			//tag添加权限检查
			if !db_perm_has(hooks_ctx.RepoId, hooks_ctx.UserId, REPOS_RULE_TAG_ADD) {
				w.Write([]byte("*** 您没有权限添加tag"))
				return
			}

			//检查tag命名是否合法
			tag_names := db_config_get(hooks_ctx.RepoId, "tag_names")
			if !MatchPartten(tag_names, name, true) {
				w.Write([]byte("*** tag命名不合法"))
				return
			}
		} else if newrev_type == "delete" {
			//tag修改权限检查
			if !db_perm_has(hooks_ctx.RepoId, hooks_ctx.UserId, REPOS_RULE_TAG_EDIT) {
				w.Write([]byte("*** 您没有权限删除tag"))
				return
			}
		} else {
			w.Write([]byte("*** newrev_type 错误"))
			return
		}

		//提交日志
		db_commit_log_add(hooks_ctx.UserId, hooks_ctx.RepoId, refname, oldrev, newrev)
	} else {
		w.Write([]byte("*** 操作不充许"))
		return
	}

	w.Write([]byte("ok"))
}

func hooks_update_change(ctx *Context) {
	hooks_update := filepath.Join(ctx.RepoDir, "/hooks/update")

	update, err := Asset("res/hook_update.tpl")
	if err != nil {
		panic(err)
	}

	update = bytes.Replace(update, []byte("{{$.Sid}}"), []byte(sid), 1)
	update = bytes.Replace(update, []byte("{{$.HookUrl}}"), []byte(hook_update_url), 1)

	err = ioutil.WriteFile(hooks_update, update, 0777)
	if err != nil {
		panic(err)
	}
}
