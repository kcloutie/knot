package main

import (
	"os"

	"github.com/kcloutie/knot/pkg/cmd/knot"
	"github.com/kcloutie/knot/pkg/params"
)

func main() {
	cliParams := params.New()
	cli := knot.Root(cliParams)

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
