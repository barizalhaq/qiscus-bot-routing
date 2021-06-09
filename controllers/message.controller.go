package controllers

import (
	"bot-routing-engine/controllers/services"
	"bot-routing-engine/entities/viewmodel"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type messageController struct {
	layerService   services.LayerService
	requestService services.RequestService
	messageService services.MessageService
	roomService    services.RoomService
}

func NewMessageController(layerService services.LayerService, requestService services.RequestService, messageService services.MessageService, roomService services.RoomService) *messageController {
	return &messageController{layerService, requestService, messageService, roomService}
}

func (controller *messageController) MessageReceived(ctx echo.Context) error {
	reqBody, err := controller.requestService.ValidateRequest(ctx, new(viewmodel.WebhookRequest))
	if err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
	}

	drafts, err := controller.messageService.Determine(reqBody)
	if err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
	}

	for _, draft := range drafts {
		if !draft.Layer.Handover && !draft.Layer.Resolve {
			err = controller.roomService.SendBotMessage(draft.Room.Payload.Room.ID, draft.Message)
			if err != nil {
				return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
			}
		}

		if draft.Layer.Resolve && !draft.Layer.Handover {
			if len(draft.Layer.Message) > 0 {
				err = controller.roomService.SendBotMessage(draft.Room.Payload.Room.ID, draft.Message)
				if err != nil {
					return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
				}
			}
			qismoRoomInfo, err := controller.roomService.QismoRoomInfo(draft.Room.Payload.Room.ID)
			if err != nil {
				return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
			}

			err = controller.roomService.Resolve(draft.Room.Payload.Room.ID, strconv.Itoa(qismoRoomInfo.Data.ID))
			if err != nil {
				return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
			}
		}

		if draft.Layer.Handover {
			err := controller.roomService.Handover(draft.Room.Payload.Room.ID)
			if err != nil {
				return ctx.JSON(http.StatusUnprocessableEntity, viewmodel.ErrorResponse{Message: err.Error()})
			}
		}
	}

	return ctx.String(http.StatusOK, "")
}
