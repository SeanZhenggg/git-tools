package main

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type branch struct {
	Project string
	Matches []string
}

func branchAction(ctx *cli.Context) error {
	user := ctx.String("name")
	if len(user) == 0 {
		return fmt.Errorf("branch name required")
	}

	dir := ctx.String("dir")

	err := branchRun(user, dir)
	if err != nil {
		return err
	}

	return nil
}

func branchFlags() []cli.Flag {
	root, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	return []cli.Flag{
		&cli.StringFlag{
			Name:    "name",
			Aliases: []string{"n"},
			Value:   "",
			Usage:   "git branch name(will use blurring search)",
		},
		&cli.StringFlag{
			Name:    "dir",
			Aliases: []string{"d"},
			Value:   root,
			Usage:   "parent directory to start recursively searching for *.git files",
		},
	}
}

func branchRun(user string, dir string) error {
	branches, err := getGitBranches(dir, user)
	if err != nil {
		return fmt.Errorf("error when getGitBranches: %w", err)
	}

	builder := strings.Builder{}
	for _, br := range branches {
		builder.WriteString(br.Project + " : \n")
		for _, branchMatch := range br.Matches {
			builder.WriteString("\t" + branchMatch)
		}
		builder.WriteString("\n")
	}

	fmt.Print(builder.String())

	return nil
}

func getGitBranches(dir, name string) ([]branch, error) {
	var branches []branch
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == ".git" {
			b, err := getBranches(path, name)
			if err != nil {
				return err
			}

			if len(b) == 0 {
				return nil
			}

			// https://stackoverflow.com/questions/27553274/unmarshal-xml-array-in-golang-only-getting-the-first-element
			//https://yourbasic.org/golang/list-files-in-directory/
			d := bytes.NewBuffer(b)
			var c branch
			c.Project = getParentDir(path)

			for {
				b, err := d.ReadBytes(byte('\n'))
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				c.Matches = append(c.Matches, string(b))
			}
			branches = append(branches, c)

		}

		return nil
	})

	return branches, err
}

func getBranches(path, name string) ([]byte, error) {
	cmd := exec.Command("git", "branch", "--all", "--list", "*"+name+"*")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}
