package main

import (
	"fmt"
	"os"

	"github.com/dc0d/argify"
	"github.com/dc0d/config/iniconfig"
	"github.com/urfave/cli"
)

func main() {
	if err := iniconfig.New().Load(&conf); err != nil {
		// this error does not help much, unless we explicitly need it
		// in which case it should be handled properly
	}

	app := cli.NewApp()
	setAppInfo(app)
	addCommands(app)
	argify.NewArgify().Build(app, &conf)

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func addCommands(app *cli.App) {
	app.Action = cmdApp
	app.Commands = append(app.Commands, cli.Command{
		Name:   "create",
		Action: cmdCreate,
	})
}

func setAppInfo(app *cli.App) {
	app.Version = "0.0.1"
	app.Author = "dc0d"
	app.Copyright = "dc0d"
	app.Name = "ggen"
	app.Usage = ""
}
