package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()

	info(app)
	commands(app)
	action(app)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func info(app *cli.App) {
	app.Name = "git-tools"
	app.Usage = "useful git tools when dealing with large project structures"
	app.Authors = []*cli.Author{{Name: "btp_sean", Email: "s410585038@gmail.com"}}
	app.Version = "0.0.1"
}

func commands(app *cli.App) {
	app.Commands = cli.Commands{
		{
			Name:   "history",
			Usage:  "log all commit histories",
			Flags:  logHistoryFlags(),
			Action: logHistoryAction,
		},
		{
			Name:   "branch",
			Usage:  "log all branches from local and remote",
			Flags:  branchFlags(),
			Action: branchAction,
		},
	}
}

func action(app *cli.App) {
	app.Action = func(ctx *cli.Context) error {
		return cli.ShowAppHelp(ctx)
	}
}
