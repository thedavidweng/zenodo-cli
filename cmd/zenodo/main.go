package main

import (
	"errors"
	"os"

	"github.com/thedavidweng/zenodo-cli/internal/cli"
	"github.com/thedavidweng/zenodo-cli/internal/model"
)

func main() {
	if err := cli.Execute(); err != nil {
		var cmdErr *model.CommandError
		if errors.As(err, &cmdErr) {
			os.Exit(model.ExitCode(cmdErr.Code))
		}
		os.Exit(1)
	}
}
