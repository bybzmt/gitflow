package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/http/cgi"
	"path/filepath"
	"strings"
)

type Context struct {
	RepoId   int
	RepoName string
	RepoDir  string
	UserId   int
	UserName string
}

type Page func(w http.ResponseWriter, r *http.Request, ctx *Context)

var gitBin = flag.String("git", "git", "git path")
var root = flag.String("repo", "./", "repositories path")
var dbdsn = flag.String("dsn", "root:123456@tcp(127.0.0.1:3306)/gitflow", "database dsn")
var addr = flag.String("addr", ":80", "listen on ip:port")

func main() {
	flag.Parse()

	http.HandleFunc("/__gitflow__/users", page_admin_users)
	http.HandleFunc("/__gitflow__/useradd", page_admin_useradd)
	http.HandleFunc("/__gitflow__/useradd_do", page_admin_useradd_do)
	http.HandleFunc("/__gitflow__/useredit", page_admin_useredit)
	http.HandleFunc("/__gitflow__/useredit_do", page_admin_useredit_do)
	http.HandleFunc("/__gitflow__/userdel", page_admin_userdel)
	http.HandleFunc("/__gitflow__/userdel_do", page_admin_userdel_do)
	http.HandleFunc("/__gitflow__/repoadd", page_admin_repoadd)
	http.HandleFunc("/__gitflow__/repoadd_do", page_admin_repoadd_do)
	http.HandleFunc("/__gitflow__/repoedit", page_admin_repoedit)
	http.HandleFunc("/__gitflow__/repoedit_do", page_admin_repoedit_do)
	http.HandleFunc("/__gitflow__/repodel", page_admin_repodel)
	http.HandleFunc("/__gitflow__/repodel_do", page_admin_repodel_do)
	http.HandleFunc("/__gitflow__/res/", page_res)
	http.HandleFunc("/favicon.ico", page_favicon)

	http.HandleFunc("/__hooks__/update", page_hooks_update)

	//git服务
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer catch_error(w, r)

		if r.URL.Path == "/" {
			page_admin_index(w, r)
			return
		}

		services := []struct {
			Pattern string
			Call    Page
		}{
			{"_gitflow_branch_getall", page_branch_getall},
			{"_gitflow_branch_getinfo", page_branch_getinfo},
			{"_gitflow_branch_has", page_branch_has},
			{"_gitflow_branch_add", page_branch_add},
			{"_gitflow_branch_delete", page_branch_delete},
			{"_gitflow_branch_status", page_branch_status},
			{"_gitflow_branch_merged", page_branch_merged},
			{"_gitflow_history_branchs", page_history_branchs},
			{"_gitflow_get_config", page_get_config},
		}

		tmps := strings.SplitN(strings.Trim(filepath.FromSlash(r.URL.Path), "/"), "/", 2)
		repo_name := tmps[0]
		repo_dir := filepath.Join(*root, tmps[0])

		//查找git库id
		repo_id := db_find_repo_id(repo_name)
		if repo_id < 1 {
			http.NotFound(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		var user_id int
		if ok {
			user_id = userAuth(repo_id, user, pass)
		}

		ctx := &Context{
			RepoId:   repo_id,
			RepoName: repo_name,
			RepoDir:  repo_dir,
			UserId:   user_id,
			UserName: user,
		}

		for _, tmp := range services {
			if strings.HasSuffix(r.URL.Path, tmp.Pattern) {

				if user_id < 1 {
					//w.Header().Add("WWW-Authenticate", "Basic realm=\"USER LOGIN\"")
					//http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

					res := Result{}
					res.Code = 1
					res.Msg = "认证失败"
					json.NewEncoder(w).Encode(res)
					return
				}

				tmp.Call(w, r, ctx)
				return
			}
		}

		if user_id < 1 {
			w.Header().Add("WWW-Authenticate", "Basic realm=\"USER LOGIN\"")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		page_git(w, r, ctx)
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func page_git(w http.ResponseWriter, r *http.Request, ctx *Context) {
	//上传
	if strings.HasSuffix(r.URL.Path, "/git-receive-pack") {
		hooks_start(ctx)
		defer hooks_end()

		//动态改变钩子
		hooks_update_change(ctx)
	}

	bin := "/usr/bin/git"

	var cgih cgi.Handler
	cgih = cgi.Handler{
		Path: bin,
		Dir:  *root,
		Root: "/",
		Args: []string{"http-backend"},
		Env: []string{
			"GIT_PROJECT_ROOT=" + *root,
			"GIT_HTTP_EXPORT_ALL=1",
		},
	}
	cgih.ServeHTTP(w, r)
}
