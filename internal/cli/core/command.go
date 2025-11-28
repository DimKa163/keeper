package core

import "github.com/spf13/cobra"

type CommandBuilder interface {
	Build() (*cobra.Command, error)
}
