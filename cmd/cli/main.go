package main

import (
	"fmt"
	"log"

	"github.com/DimKa163/keeper/app/cmd"
	"github.com/DimKa163/keeper/internal/cli"
)

func main() {
	app, err := cmd.New()
	if err != nil {
		log.Fatal(err)
	}
	app.AddCommand("create-login-pass", cli.AddLoginPassCommand())
	if err := app.Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}
