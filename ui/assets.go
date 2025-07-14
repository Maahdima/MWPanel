package ui

import "embed"

//go:embed dist/*
var fs embed.FS

func GetUiAssets() embed.FS {
	return fs
}
