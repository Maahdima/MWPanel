package http

import (
	"errors"
	"io"
	"io/fs"
	gohttp "net/http"

	"github.com/labstack/echo/v4"
)

type UIController struct {
	uiAssetsFs             fs.FS
	staticDirectoryHandler echo.HandlerFunc
}

func NewUiController(fs fs.FS) *UIController {
	return &UIController{
		uiAssetsFs:             fs,
		staticDirectoryHandler: echo.StaticDirectoryHandler(fs, false),
	}
}

func (c *UIController) Serve(ctx echo.Context) error {
	if err := c.staticDirectoryHandler(ctx); err != nil {
		f, err := c.uiAssetsFs.Open("index.html")
		if err != nil {
			return echo.ErrNotFound
		}
		defer f.Close()

		fi, _ := f.Stat()
		ff, ok := f.(io.ReadSeeker)
		if !ok {
			return errors.New("file does not implement io.ReadSeeker")
		}
		gohttp.ServeContent(ctx.Response(), ctx.Request(), fi.Name(), fi.ModTime(), ff)
		return nil
	}

	return nil
}

func SetupMwpUI(app *echo.Echo, uiAssetsFs fs.FS) {
	app.GET("/*", NewUiController(uiAssetsFs).Serve)
}
