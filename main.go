// Package main is the entry point to the program.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/leonhfr/cete/cmd"
	"github.com/leonhfr/cete/static"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cmd.Execute(ctx, static.FileSystem)
}
