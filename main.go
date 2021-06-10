package main

import (
	"bot-routing-engine/entities"
	"bot-routing-engine/routes"
	"bot-routing-engine/utils/logger"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var loggers map[string]*log.Logger

func init() {
	loggers = make(map[string]*log.Logger)
	loggers["inbound"] = logger.InitLogger("./logs/inbound", "")
	loggers["outbond"] = logger.InitLogger("./logs/outbond", "")

	godotenv.Load(".env")
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyDump(func(ctx echo.Context, reqBody, resBody []byte) {
		loggers["inbound"].Printf("REQUEST: %s %s %s Header: %s Payload: %s\n", ctx.Request().RemoteAddr, ctx.Request().Method, ctx.Request().URL, ctx.Request().Header, string(reqBody))
		loggers["inbound"].Printf("RESPONSE: %d %s", ctx.Response().Status, string(resBody))
	}))

	multichannel := entities.NewMultichannel(
		os.Getenv("MULTICHANNEL_APP_ID"),
		os.Getenv("MULTICHANNEL_ADMIN_EMAIL"),
		os.Getenv("MULTICHANNEL_SECRET"),
		os.Getenv("MULTICHANNEL_TOKEN"),
	)

	route := routes.New(e, multichannel, loggers["outbond"])
	route.RegisterRoute()

	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}
