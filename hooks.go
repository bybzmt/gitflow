package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var hooks_lock sync.Mutex
var sid = ""
var hooks_ctx *Context

func hooks_start(ctx *Context) {
	hooks_lock.Lock()
	tmp := rand.Int63()
	sid = strconv.FormatInt(tmp, 16)
	hooks_ctx = ctx
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

		status := db_branch_get_status(hooks_ctx.RepoId, name)

		if status == 1 {
			if !db_perm_has(hooks_ctx.RepoId, hooks_ctx.UserId, 3) {
				w.Write([]byte("branch no perm"))
				return
			}
		} else if status == 2 {
			if !db_perm_has(hooks_ctx.RepoId, hooks_ctx.UserId, 2) {
				w.Write([]byte("branch no perm"))
				return
			}
		} else if status == 3 {
			w.Write([]byte("branch is merge apply"))
			return
		} else if status == 4 {
			w.Write([]byte("branch is merged"))
			return
		} else if status == 5 {
			w.Write([]byte("branch is deleted"))
			return
		} else {
			w.Write([]byte("branch error"))
			return
		}

		//更新分支提交
		db_branch_update_commit(hooks_ctx.RepoId, name, newrev)

		//提交日志
		db_commit_log_add(hooks_ctx.UserId, hooks_ctx.RepoId, refname, oldrev, newrev)

	} else if strings.HasPrefix(refname, "refs/tags/") {
		if !db_perm_has(hooks_ctx.RepoId, hooks_ctx.UserId, 3) {
			w.Write([]byte("tags no perm"))
			return
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

	update := `#!/bin/bash
refname=$(echo "$1" |tr -d '\n' |od -An -tx1|tr ' ' %)
oldrev=$(echo "$2" |tr -d '\n' |od -An -tx1|tr ' ' %)
newrev=$(echo "$3" |tr -d '\n' |od -An -tx1|tr ' ' %)
sid="{{sid}}"
ok=$(curl -s -d "sid=${sid}&refname=${1}&oldrev=${2}&newrev=${3}" "http://127.0.0.1:8088/__hooks__/update")
if [[ $ok != "ok" ]]; then
	echo $ok
	exit 1
fi
exit 0
`
	fh, err := os.OpenFile(hooks_update, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	update = strings.Replace(update, "{{sid}}", sid, 1)

	_, err = fh.WriteString(update)
	if err != nil {
		panic(err)
	}
}
