package main

import (
	"fmt"
	"github.com/DimKa163/keeper/app/cmd"
)

var (
	Version string
	Commit  string
	Date    string
)

func main() {
	app, err := cmd.New(Version, Commit, Date)
	if err != nil {
		panic(err)
	}
	if err = app.RegisterCommands(); err != nil {
		panic(err)
	}
	if err = app.Execute(); err != nil {
		fmt.Println("Error: ", err)
	}
}
