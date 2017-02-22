package main

import (
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"regexp/syntax"
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

func git_exec(repo_dir string, arg ...string) ([]byte, error) {
	err := os.Chdir(repo_dir)
	if err != nil {
		return nil, err
	}

	out, err := exec.Command(*gitBin, arg...).CombinedOutput()

	//命令执行日志
	log.Println("cmd:", repo_dir, gitBin, strings.Join(arg, " "), "[", err, "]", string(out))

	if err != nil {
		return out, err
	}

	return out, nil
}

func splitLines(text string) (out []string) {
	//处理dot
	regx, err := syntax.Parse("\r|\n", syntax.PerlX|syntax.MatchNL|syntax.UnicodeGroups)
	if err != nil {
		log.Println("regex error:", err.Error())
		return nil
	}

	tmps := regexp.MustCompile(regx.String()).Split(text, -1)
	for _, tmp := range tmps {
		tmp = strings.Trim(tmp, " \r\n")
		if tmp != "" {
			out = append(out, tmp)
		}
	}

	return
}
