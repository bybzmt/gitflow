package main

import (
	"flag"
	"log"
	"net/http"
)

var gitBin = flag.String("git", "/usr/bin/git", "git bin path")
var root = flag.String("repos", "./", "repositories path")
var dbdsn = flag.String("dsn", "root:123456@tcp(127.0.0.1:3306)/gitflow", "database dsn")
var addr = flag.String("addr", ":80", "listen on ip:port")

func main() {
	flag.Parse()

	server_init()
	config.ProjectRoot = *root
	config.GitBinPath = *gitBin

	//管理界面
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
	http.HandleFunc("/favicon.ico", page_file("favicon.ico"))
	http.HandleFunc("/", page_file("index.html"))

	//钩子回调
	http.HandleFunc("/__hooks__/update", page_hooks_update)

	//git服务
	http.Handle("/repos/", http.StripPrefix("/repos/", http.HandlerFunc(requestHandler)))

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func server_init() {
	if db_is_empty() {
		log.Println("数据库为空，自动创建相关表.")
		db_init_tables()
		log.Println("创建相关表完成.")
		log.Println("目前所有用户为空，请注册用户，第1个注册的用户为管理员.")
	}
}
