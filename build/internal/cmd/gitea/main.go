package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

func main() {
	var giteaURL = flag.String("gitea-url", "", "Address for target gitea service")
	var testdataDir = flag.String("testdata-dir", "", "Directory path to testdata")
	flag.Parse()

	fatalOnError := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	if *giteaURL == "" {
		log.Fatal("Must supply non-empty --gitea-url flag value.")
	}

	fmt.Println("Configuring Gitea at", *giteaURL)

	resp, err := http.Post(*giteaURL, "application/x-www-form-urlencoded", strings.NewReader(giteaSetupForm))
	fatalOnError(err)

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Expected Status OK; Found %d", resp.StatusCode)
	}

	var cli *gitea.Client
	for i := 0; i < 10; i++ {
		cli, err = gitea.NewClient(*giteaURL, gitea.SetBasicAuth("root", "password"))
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
	}

	origin, _, err := cli.CreateRepo(gitea.CreateRepoOption{
		Name:          "features",
		DefaultBranch: "main",
	})
	fatalOnError(err)

	workdir := memfs.New()

	fmt.Println("Creating Repository from", *testdataDir)

	repo, err := git.InitWithOptions(memory.NewStorage(), workdir, git.InitOptions{
		DefaultBranch: "main",
	})
	fatalOnError(err)

	repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{origin.CloneURL},
	})

	tree, err := repo.Worktree()
	fatalOnError(err)

	dir := os.DirFS(*testdataDir)
	fatalOnError(err)

	// copy testdata into target tmp dir
	err = fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return workdir.MkdirAll(path, 0755)
		}

		contents, err := fs.ReadFile(dir, path)
		if err != nil {
			return err
		}

		fi, err := workdir.Create(path)
		if err != nil {
			return err
		}

		_, err = fi.Write(contents)
		if err != nil {
			return err
		}

		return fi.Close()
	})
	fatalOnError(err)

	err = tree.AddWithOptions(&git.AddOptions{All: true})
	fatalOnError(err)

	_, err = tree.Commit("feat: add entire contents", &git.CommitOptions{
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	fatalOnError(err)

	fmt.Println("Pushing to", origin.CloneURL)
	repo.Push(&git.PushOptions{
		Auth:       &githttp.BasicAuth{Username: "root", Password: "password"},
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"main:refs/heads/main"},
	})
}

const giteaSetupForm = "db_type=sqlite3&db_host=localhost%3A3306&db_user=root&db_passwd=&db_name=gitea&ssl_mode=disable&db_schema=&charset=utf8&db_path=%2Fdata%2Fgitea%2Fgitea.db&app_name=Gitea%3A+Git+with+a+cup+of+tea&repo_root_path=%2Fdata%2Fgit%2Frepositories&lfs_root_path=%2Fdata%2Fgit%2Flfs&run_user=git&domain=localhost&ssh_port=22&http_port=3000&app_url=http%3A%2F%2Flocalhost%3A3000%2F&log_root_path=%2Fdata%2Fgitea%2Flog&smtp_addr=&smtp_port=&smtp_from=&smtp_user=&smtp_passwd=&enable_federated_avatar=on&enable_open_id_sign_in=on&enable_open_id_sign_up=on&default_allow_create_organization=on&default_enable_timetracking=on&no_reply_address=noreply.localhost&password_algorithm=pbkdf2&admin_name=root&admin_passwd=password&admin_confirm_passwd=password&admin_email=dev%40flipt.io"
