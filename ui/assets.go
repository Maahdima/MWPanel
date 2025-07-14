package ui

import "embed"

//go:embed dist/*
var fs embed.FS

func GetUIAssets() embed.FS {
	return fs
}
