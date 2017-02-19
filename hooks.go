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

	log.Println(sid, oldrev, newrev, refname)

	if sid != _sid {
		w.Write([]byte("hooks sid fail"))
		return
	}

	db := db_open()
	defer db.Close()

	if strings.HasPrefix(refname, "refs/heads/") {
		name := refname[len("refs/heads/"):]

		//TODO 写权限检查
		//TODO 锁定分支检查
		//TODO 锁定分支权限检查
		//TODO 分支命名规则检查
		//TODO 强型推送检查

		branch_names := db_config_get(hooks_ctx.RepoId, "branch_names")
		if branch_names != "" {
			if name != branch_names {
				w.Write([]byte("branch no perm"))
				return
			}
		}

		//提交日志
		db_commit_log_add(hooks_ctx.UserId, hooks_ctx.RepoId, refname, oldrev, newrev)

	} else if strings.HasPrefix(refname, "refs/tags/") {
		name := refname[len("refs/tags/"):]

		//TODO tag添加权限检查
		//TODO tag修改权限检查

		if !db_perm_has(hooks_ctx.RepoId, hooks_ctx.UserId, 3) {
			w.Write([]byte("tags no perm"))
			return
		}

		tag_names := db_config_get(hooks_ctx.RepoId, "tag_names")
		if tag_names != "" {
			if name != tag_names {
				w.Write([]byte("tag no perm"))
				return
			}
		}

		//提交日志
		db_commit_log_add(hooks_ctx.UserId, hooks_ctx.RepoId, refname, oldrev, newrev)
	} else {
		w.Write([]byte("action not allow"))
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
