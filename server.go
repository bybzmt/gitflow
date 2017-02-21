package main

import (
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
	BaseUrl  string
}

type Page func(w http.ResponseWriter, r *http.Request, ctx *Context)

var gitBin = flag.String("git", "/usr/bin/git", "git bin path")
var root = flag.String("repos", "./", "repositories path")
var dbdsn = flag.String("dsn", "root:123456@tcp(127.0.0.1:3306)/gitflow", "database dsn")
var addr = flag.String("addr", ":80", "listen on ip:port")

func main() {
	flag.Parse()

	//server_init()

	//管理界面
	http.HandleFunc("/", page_admin_index)
	http.HandleFunc("/admin/users", page_admin_users)
	http.HandleFunc("/admin/useradd", page_admin_useradd)
	http.HandleFunc("/admin/useradd_do", page_admin_useradd_do)
	http.HandleFunc("/admin/useredit", page_admin_useredit)
	http.HandleFunc("/admin/useredit_do", page_admin_useredit_do)
	http.HandleFunc("/admin/userdel", page_admin_userdel)
	http.HandleFunc("/admin/userdel_do", page_admin_userdel_do)
	http.HandleFunc("/admin/repoadd", page_admin_repoadd)
	http.HandleFunc("/admin/repoadd_do", page_admin_repoadd_do)
	http.HandleFunc("/admin/repoedit", page_admin_repoedit)
	http.HandleFunc("/admin/repoedit_do", page_admin_repoedit_do)
	http.HandleFunc("/admin/repodel", page_admin_repodel)
	http.HandleFunc("/admin/repodel_do", page_admin_repodel_do)
	//静态资源
	http.HandleFunc("/res/", page_res)
	http.HandleFunc("/favicon.ico", page_favicon)
	//钩子回调
	http.HandleFunc("/__hooks__/update", page_hooks_update)
	//git服务
	//http.Handle("/repos/", http.StripPrefix("/repos/", http.HandlerFunc(page_repos)))
	http.HandleFunc("/repos/", page_repos)

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func page_repos(w http.ResponseWriter, r *http.Request) {
	defer catch_error(w, r)

	var repo_name, repo_dir string
	tmps := strings.SplitN(strings.Trim(filepath.FromSlash(r.URL.Path), "/"), "/", 3)
	if len(tmps) > 2 {
		repo_name = tmps[1]
		repo_dir = filepath.Join(*root, tmps[1])
	}

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
		BaseUrl:  "http://" + r.Host,
	}

	if user_id < 1 {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"USER LOGIN\"")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !db_perm_has(repo_id, user_id, REPOS_RULE_READ) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	page_git(w, r, ctx)
}

func page_git(w http.ResponseWriter, r *http.Request, ctx *Context) {
	//上传
	if strings.HasSuffix(r.URL.Path, "/git-receive-pack") {
		hooks_start(ctx)
		defer hooks_end()

		//动态改变钩子
		hooks_update_change(ctx)
	}

	var cgih cgi.Handler
	cgih = cgi.Handler{
		Path: *gitBin,
		Dir:  *root,
		Root: "/repos",
		Args: []string{"http-backend"},
		Env: []string{
			"GIT_HTTP_EXPORT_ALL=",
			"GIT_PROJECT_ROOT=" + *root,
		},
	}
	cgih.ServeHTTP(w, r)
}
