package main

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"regexp/syntax"
	"strings"
)

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
