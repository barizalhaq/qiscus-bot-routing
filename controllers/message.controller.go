package controllers

import (
	"bot-routing-engine/controllers/services"
	"bot-routing-engine/entities/viewmodel"
	"net/http"

	"github.com/labstack/echo/v4"
)

type messageController struct {
	layerService   services.LayerService
	requestService services.RequestService
	messageService services.MessageService
}

func NewMessageController(layerService services.LayerService, requestService services.RequestService, messageService services.MessageService) *messageController {
	return &messageController{layerService, requestService, messageService}
}

func (controller *messageController) MessageReceived(ctx echo.Context) error {
	reqBody, err := controller.requestService.ValidateRequest(ctx, new(viewmodel.WebhookRequest))
	if err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
	}

	layer, room, err := controller.messageService.Determine(reqBody)
	if err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
	}

	if !layer.Handover {
		err = controller.messageService.SendBotMessage(room.Results.Rooms[0].ID, layer.Message)
	}

	if err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
	}

	return ctx.String(http.StatusOK, "")
}
