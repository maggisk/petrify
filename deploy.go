package main

import "fmt"
import "os"
import "strings"
import "io/ioutil"
import "path/filepath"
import "time"
import "syscall"
import "golang.org/x/crypto/ssh/terminal"
import "gopkg.in/src-d/go-git.v4"
import gitconfig "gopkg.in/src-d/go-git.v4/config"
import "gopkg.in/src-d/go-git.v4/plumbing/object"
import "gopkg.in/src-d/go-git.v4/plumbing/transport"
import "gopkg.in/src-d/go-git.v4/plumbing/transport/http"

func Deploy(config *Config) {
	config.ValidateForDeploy()
	if config.DeployToGithub != "" {
		deployToGithub(config)
	}
	// TODO: more deployment options
}

func deployToGithub(config *Config) {
	// prompt for password when using https
	var auth *http.BasicAuth
	if strings.HasPrefix(config.DeployToGithub, "https:") {
		auth = getAuth("github", config.GithubUsername, "GITHUB_USERNAME", config.GithubPassword, "GITHUB_PASSWORD")
	}

	dir, err := ioutil.TempDir("", "petrify-deploy-")
	checkError(err)
	defer os.RemoveAll(dir)

	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      config.DeployToGithub,
		Progress: os.Stdout,
		Auth:     auth,
	})

	if err != transport.ErrEmptyRemoteRepository {
		checkError(err)
	} else {
		repo, err = git.PlainInit(dir, false)
		checkError(err)
		_, err = repo.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{config.DeployToGithub},
		})
		checkError(err)
	}

	tree, err := repo.Worktree()
	checkError(err)

	tree.RemoveGlob("*")
	CopyDir(config.BuildDir, dir)

	if len(config.Path404) > 0 && config.Path404 != "/404.html" {
		// TODO: github expects the filename to be 404.html
	}

	if len(config.CNAME) > 0 {
		checkError(ioutil.WriteFile(filepath.Join(dir, "CNAME"), []byte(config.CNAME+"\n"), FILE_BITMASK))
		tree.Add("CNAME")
	}

	files, err := ioutil.ReadDir(config.BuildDir)
	checkError(err)
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			tree.Add(file.Name())
		}
	}

	status, err := tree.Status()
	checkError(err)

	if len(status) == 0 {
		info("Nothing to commit")
		return
	}

	_, err = tree.Commit(time.Now().Format("petrify commit @ "+time.ANSIC), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Petrify",
			Email: "",
			When:  time.Now(),
		},
	})
	checkError(err)

	checkError(repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		Auth:       auth,
	}))
}

func getAuth(siteName, username, usernameEnv, password, passwordEnv string) *http.BasicAuth {
	if username == "" {
		username = os.Getenv(usernameEnv)
	}
	if username == "" {
		fmt.Printf("%s username: ", siteName)
		username = strings.TrimRight(ReadLine(), "\r\n")
	}

	if password == "" {
		password = os.Getenv(passwordEnv)
	}
	if password == "" {
		fmt.Printf("%s password: ", siteName)
		passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		checkError(err)
		password = string(passwordBytes)
	}

	return &http.BasicAuth{
		Username: username,
		Password: password,
	}
}
