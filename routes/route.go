package routes

import (
	"bot-routing-engine/controllers"
	"bot-routing-engine/controllers/services"
	"bot-routing-engine/entities"
	"bot-routing-engine/repositories"
	"log"

	"github.com/labstack/echo/v4"
)

type route struct {
	e             *echo.Echo
	Multichannel  *entities.Multichannel
	outbondLogger *log.Logger
}

func New(e *echo.Echo, multichannel *entities.Multichannel, outbondLogger *log.Logger) *route {
	return &route{e, multichannel, outbondLogger}
}

func (r *route) RegisterRoute() {
	apiGroup := r.e.Group("/api/v1")

	messageGroup := apiGroup.Group("/message")
	appGroup := apiGroup.Group("/app")

	roomRepo := repositories.NewRoomRepository(r.Multichannel, r.outbondLogger)
	mulchanRepo := repositories.NewMultichannelRepository(r.Multichannel, r.outbondLogger)

	layerService := services.NewLayerService()
	requestService := services.NewRequestService()
	roomService := services.NewRoomService(mulchanRepo, roomRepo)
	messageService := services.NewMessageService(mulchanRepo, *roomService)

	messageController := controllers.NewMessageController(layerService, requestService, messageService, roomService)
	uploadController := controllers.NewUploadController()

	messageGroup.POST("/received", messageController.MessageReceived)
	appGroup.POST("/upload", uploadController.Upload)
}
