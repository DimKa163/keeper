package cli

import (
	"fmt"
)

func AddLoginPassCommand() CommandHandler {
	return func(ctx *CLI) error {
		consLin := ctx.Console()
		login, err := consLin.ReadRequiredString("input login")
		if err != nil {
			return err
		}
		pass, err := consLin.ReadRequiredString("input password")
		if err != nil {
			return err
		}
		fmt.Println("login:", login)
		fmt.Println("password:", pass)
		return nil
	}
}
