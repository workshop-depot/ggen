package main

import (
	"os"

	"github.com/urfave/cli"
)

func cmdApp(*cli.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		logerr.Println(err)
		return nil
	}

	process(wd)

	return nil
}
