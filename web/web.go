package web

import (
	"embed"
	"io/fs"
)

//go:embed static
var static embed.FS

func Static() fs.FS {
	sub, _ := fs.Sub(static, "static")
	return sub
}
