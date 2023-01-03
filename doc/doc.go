package doc

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed swagger/*
var swggerEFS embed.FS

func SwaggerFS() http.FileSystem {
	swggerEFS, err := fs.Sub(swggerEFS, "swagger")
	if err != nil {
		panic(err)
	}
	return http.FS(swggerEFS)
}
