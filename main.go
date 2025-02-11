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
	"strings"
	"time"
)

type commit struct {
	Author  string `xml:"author"`
	Project string `xml:"project"`
	Date    string `xml:"date"`
	Message string `xml:"message"`
}

func info(app *cli.App) {
	app.Name = "git-history-tools"
	app.Usage = "reports git history"
	app.Authors = []*cli.Author{{Name: "btp_sean", Email: "s410585038@gmail.com"}}
	app.Version = "0.0.0"
}

func flags(app *cli.App) {
	root, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	app.Flags = []cli.Flag{
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

func commands(app *cli.App) {
	app.Action = func(ctx *cli.Context) error {
		user := ctx.String("user")
		if len(user) == 0 {
			log.Println("user is not defined, use default git config user instead")
			cmd := exec.Command("git", "config", "user.name")
			out, err := cmd.Output()
			if err != nil {
				return fmt.Errorf("when getting name fron git config - %v", err)
			}

			user = string(out)
			if len(user) == 0 {
				return fmt.Errorf("no user name found in git config")
			}
		}

		dir := ctx.String("dir")
		afterDate := ctx.Timestamp("after")

		err := run(user, dir, afterDate)
		if err != nil {
			return err
		}

		log.Println("done")
		return nil
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

	err = os.WriteFile("history.json", indent, 0776)
	if err != nil {
		return fmt.Errorf("error when write history.json: %w", err)
	}
	return nil
}

func main() {
	app := cli.NewApp()
	info(app)
	flags(app)
	commands(app)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
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

				c.Project = getParentDir(path)
				commits = append(commits, c)
			}

		}

		return nil
	})

	return commits, err
}

func getParentDir(path string) string {
	n := strings.Split(path, "/")
	if len(n) < 2 {
		return ""
	}
	return n[len(n)-2]
}

func getCommits(path, user, after string) ([]byte, error) {
	format := `<entry>
				<author>%an</author>
				<date>%cd</date>
				<message>"%B"</message>
				</entry>`
	cmd := exec.Command("git", "log", "--author="+user, "--pretty=format:"+format, "--after="+after)
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}
