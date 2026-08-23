package http

import (
	"net/http"

	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type TelegramController struct {
	bot    *service.TelegramBot
	logger *zap.Logger
}

func NewTelegramController(bot *service.TelegramBot) *TelegramController {
	return &TelegramController{
		bot:    bot,
		logger: zap.L().Named("TelegramController"),
	}
}

func (c *TelegramController) GetStatus(ctx echo.Context) error {
	status := schema.TelegramStatusResponse{}
	if c.bot != nil {
		botStatus := c.bot.Status()
		status.Enabled = botStatus.Enabled
		status.BotUsername = botStatus.BotUsername
		status.BotURL = botStatus.BotURL
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.TelegramStatusResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          status,
	})
}
