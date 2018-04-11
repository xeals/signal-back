package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	"github.com/xeals/signal-back/cmd"
	"github.com/xeals/signal-back/types"
)

var version = "0.0.0"

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("signal-back %s\nproto commit: %s\n", version, types.ProtoCommitHash)
	}

	app := cli.NewApp()
	app.CustomAppHelpTemplate = cmd.AppHelp
	app.Version = version
	app.Commands = []cli.Command{
		cmd.Format,
		cmd.Analyse,
		cmd.Extract,
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "show help",
		},
	}
	app.Action = func(c *cli.Context) error {
		cli.ShowAppHelpAndExit(c, 255)
		return nil
	}
	// app.Action = cli.ActionFunc(func(c *cli.Context) error {
	// 	// -- Logging

	// 	if c.String("log") != "" {
	// 		f, err := os.OpenFile(c.String("log"), os.O_CREATE|os.O_WRONLY, 0644)
	// 		if err != nil {
	// 			return errors.Wrap(err, "unable to create logging file")
	// 		}
	// 		logger = f
	// 	} else {
	// 		logger = os.Stderr
	// 	}
	// 	return nil
	// })

	if err := app.Run(os.Args); err != nil {
		// log.Fatalln(err)
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
