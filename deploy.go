package main

import "os"
import "strings"
import "io/ioutil"
import "path/filepath"
import "time"
import "gopkg.in/src-d/go-git.v4"
import gitconfig "gopkg.in/src-d/go-git.v4/config"
import "gopkg.in/src-d/go-git.v4/plumbing/object"
import "gopkg.in/src-d/go-git.v4/plumbing/transport"

func deployToGithub(config *Config) {
	dir, err := ioutil.TempDir("", "petrify-deploy-")
	checkError(err)
	defer os.RemoveAll(dir)

	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      config.DeployToGithub,
		Progress: os.Stdout,
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
	}))
}
