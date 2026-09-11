package http

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	gohttp "net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type UIController struct {
	uiAssetsFs             fs.FS
	staticDirectoryHandler echo.HandlerFunc
	adminPanelPath         string
}

func NewUiController(fs fs.FS, adminPanelPath string) *UIController {
	return &UIController{
		uiAssetsFs:             fs,
		staticDirectoryHandler: echo.StaticDirectoryHandler(fs, false),
		adminPanelPath:         adminPanelPath,
	}
}

func (c *UIController) Serve(ctx echo.Context) error {
	path := ctx.Request().URL.Path

	if c.isStaticAssetPath(path) {
		return c.staticDirectoryHandler(ctx)
	}

	if !c.allowsSPAFallback(path) {
		return echo.ErrNotFound
	}

	if err := c.staticDirectoryHandler(ctx); err == nil {
		return nil
	}

	return c.serveIndexHTML(ctx)
}

func (c *UIController) isStaticAssetPath(path string) bool {
	return strings.HasPrefix(path, "/assets/") ||
		strings.HasPrefix(path, "/images/") ||
		path == "/favicon.ico" ||
		path == "/favicon.svg"
}

func (c *UIController) allowsSPAFallback(path string) bool {
	if path == "/share" || strings.HasPrefix(path, "/share/") {
		return true
	}

	if c.adminPanelPath == "" {
		return true
	}

	return path == c.adminPanelPath || strings.HasPrefix(path, c.adminPanelPath+"/")
}

func (c *UIController) serveIndexHTML(ctx echo.Context) error {
	f, err := c.uiAssetsFs.Open("index.html")
	if err != nil {
		return echo.ErrNotFound
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]string{
		"adminPanelPath": c.adminPanelPath,
	})
	if err != nil {
		return err
	}

	injection := []byte("<script>window.__MWP_CONFIG__=" + string(payload) + ";</script>")
	html := raw
	if bytes.Contains(raw, []byte("</head>")) {
		html = bytes.Replace(raw, []byte("</head>"), append(injection, []byte("</head>")...), 1)
	} else {
		html = append(injection, raw...)
	}

	reader := bytes.NewReader(html)
	fi, _ := f.Stat()
	gohttp.ServeContent(ctx.Response(), ctx.Request(), fi.Name(), fi.ModTime(), reader)
	return nil
}

func SetupMwpUI(app *echo.Echo, uiAssetsFs fs.FS, adminPanelPath string) {
	app.GET("/*", NewUiController(uiAssetsFs, adminPanelPath).Serve)
}
