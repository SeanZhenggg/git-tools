package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type commit struct {
	Commit  string    `xml:"commit"`
	Author  string    `xml:"author"`
	Project string    `xml:"project"`
	Date    time.Time `xml:"date"`
	Message string    `xml:"message"`
}

func logHistoryAction(ctx *cli.Context) error {
	user := ctx.String("user")
	if len(user) == 0 {
		log.Println("user is not defined, use default git config user instead")
		cmd := exec.Command("git", "config", "user.name")
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("error when getting name fron git config - %v", err)
		}

		user = string(out)
		if len(user) == 0 {
			return fmt.Errorf("error no user name found in git config")
		}
	}

	dir := ctx.String("dir")
	afterDate := ctx.Timestamp("after")

	err := run(user, dir, afterDate)
	if err != nil {
		return err
	}

	return nil
}

func logHistoryFlags() []cli.Flag {
	root, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	return []cli.Flag{
		&cli.StringFlag{
			Name:    "user",
			Aliases: []string{"u"},
			Value:   "",
			Usage:   "git user name",
		},
		&cli.StringFlag{
			Name:    "dir",
			Aliases: []string{"d"},
			Value:   root,
			Usage:   "parent directory to start recursively searching for *.git files",
		},
		&cli.TimestampFlag{
			Name:    "after",
			Aliases: []string{"a"},
			Layout:  time.DateOnly,
			Value:   cli.NewTimestamp(time.Now().Add(-24 * time.Hour)),
			Usage:   "when to start looking at commit history",
		},
	}
}

func run(user string, dir string, afterDate *time.Time) error {
	history, err := getGitHistory(dir, user, *afterDate)
	if err != nil {
		return fmt.Errorf("error when getGitHistory: %w", err)
	}

	indent, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("error when json.MarshalIndent: %w", err)
	}

	pager := exec.Command("less")

	buffer := bytes.NewBuffer(indent)
	pager.Stderr = os.Stderr
	pager.Stdin = buffer
	pager.Stdout = os.Stdout

	err = pager.Run()
	if err != nil {
		return fmt.Errorf("error when less command execute: %w", err)
	}
	return nil
}

func getGitHistory(dir, user string, after time.Time) ([]commit, error) {
	var commits []commit
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == ".git" {
			b, err := getCommits(path, user, after.Format(time.DateTime))
			if err != nil {
				return err
			}

			if len(b) == 0 {
				log.Printf("no commits for user %s in project %s", user, getParentDir(path))
				return nil
			}

			// https://stackoverflow.com/questions/27553274/unmarshal-xml-array-in-golang-only-getting-the-first-element
			//https://yourbasic.org/golang/list-files-in-directory/
			d := xml.NewDecoder(bytes.NewBuffer(b))
			d.Strict = false // for now

			for {
				var c commit
				err := d.Decode(&c)
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				ch := []rune(c.Commit)
				c.Commit = string(ch[0:6])
				c.Project = getParentDir(path)
				commits = append(commits, c)
			}

		}

		return nil
	})

	return commits, err
}

func getCommits(path, user, after string) ([]byte, error) {
	format := `<entry>
				<commit>%H</commit>
				<author>%an</author>
				<date>%cI</date>
				<message>%B</message>
				</entry>`
	cmd := exec.Command("git", "log", "--author="+user, "--pretty=format:"+format, "--after="+after)
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}
