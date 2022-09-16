// Package static exports a file system providing access to the static assets.
package static

import (
	"embed"
	"io/fs"
)

//go:embed *
var embedded embed.FS

// FileSystem provides access to the static assets.
var FileSystem = fs.FS(embedded)
