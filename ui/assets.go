package ui

import "embed"

// Include Vite chunks generated for dynamic routes such as _id-*.js.
//
//go:embed all:dist/*
var fs embed.FS

func GetUIAssets() embed.FS {
	return fs
}
