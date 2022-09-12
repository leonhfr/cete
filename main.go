// Package main is the entry point to the program.
package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"os/signal"

	"github.com/leonhfr/cete/cmd"
)

//go:embed static
var staticEmbedded embed.FS

func main() {
	staticFS, err := fs.Sub(fs.FS(staticEmbedded), "static")
	if err != nil {
		log.Fatal("could not navigate to the static subtree", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cmd.Execute(ctx, staticFS)
}
