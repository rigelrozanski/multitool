package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

func replaceStringInFile(f os.FileInfo, dir, oldS, newS string) error {
	filename := path.Join(dir, f.Name())
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	w := strings.Replace(string(b), oldS, newS, -1)
	return ioutil.WriteFile(filename, []byte(w), f.Mode())
}

func replace(dir, oldS, newS string, depth int) error {
	if depth == 0 {
		return nil
	}
	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range dirFiles {
		name := f.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		p := path.Join(dir, name)
		if f.IsDir() {
			if name != "vendor" {
				if err := replace(p, oldS, newS, depth-1); err != nil {
					return err
				}
			}
			//TODO make suffixes optional & extensible
		} else if strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".md") {
			replaceStringInFile(f, dir, oldS, newS)
		}
	}
	return nil
}

func simpleCmd(name string, args ...string) error {
	var cmd *exec.Cmd
	cmd = exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func simpleCmdReturn(name string, args ...string) (*bytes.Buffer, error) {
	var cmd *exec.Cmd
	cmd = exec.Command(name, args...)
	out := new(bytes.Buffer)
	cmd.Stdout = out
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	return out, err
}

func gitCheckout(dir, branch string) error {
	cur, _ := os.Getwd()

	os.Chdir(dir)
	err := simpleCmd("git", "checkout", branch)
	if err != nil {
		return err
	}
	os.Chdir(cur)
	return nil
}

// ErrNotGitRepo - Error for not a git repo
var ErrNotGitRepo = fmt.Errorf("Not a git repo")

func gitGetBranch(dir string) (string, error) {
	if _, err := os.Stat(path.Join(dir, ".git")); err != nil {
		return "", ErrNotGitRepo
	}
	p := path.Join(dir, ".git", "HEAD")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}

	s := strings.TrimSpace(string(b))
	sp := strings.Split(s, " ")
	s = sp[len(sp)-1]
	branch := path.Base(s)
	return branch, nil
}

func gitPull(remote, branch string) error {
	return simpleCmd("git", "pull", remote, branch)
}

func addCommit(m string) error {
	err := simpleCmd("git", "add", "-u")
	if err != nil {
		return err
	}
	return simpleCmd("git", "commit", "-m", m)
}

func resolveRemoteRepo(remote string) (string, error) {
	var url string
	out, err := simpleCmdReturn("git", "config", "--local", "--get", "remote."+remote+".url")
	if err != nil {
		return "", err
	}
	fetchurl := out.String()
	if strings.Contains(fetchurl, "@") {
		sp := strings.Split(fetchurl, ":")
		if len(sp) != 2 {
			return "", fmt.Errorf("improper ssh address")
		}
		resource := sp[1]
		ssh := sp[0]
		sshsp := strings.Split(ssh, "@")
		host := sshsp[1]
		url = path.Join(host, resource)
	} else {
	}
	url = strings.TrimSpace(url)
	return url, nil
}

func resolveLocalRepo(wd string) (string, error) {
	repo := strings.TrimPrefix(wd, GoSrc)
	repo = strings.Trim(repo, "/")
	dirs := strings.Split(repo, "/")
	if len(dirs) < 3 {
		return "", fmt.Errorf("Not a valid got repo")
	}
	hub := dirs[0]
	auth := dirs[1]
	proj := dirs[2]
	repo = path.Join(hub, auth, proj)
	return repo, nil
}
