package main

import (
	"fmt"
	"log"

	"github.com/DimKa163/keeper/app/cmd"
)

func main() {
	app, err := cmd.New()
	if err != nil {
		log.Fatal(err)
	}
	app.RegisterCommands()
	if err := app.Execute(); err != nil {
		fmt.Println("Error: ", err)
	}
}
