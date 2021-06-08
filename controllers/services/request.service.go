package services

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type RequestService interface {
	ValidateRequest(ctx echo.Context, request interface{}) (interface{}, error)
}

type requestService struct{
	validator *validator.Validate
}

func NewRequestService() *requestService {
	return &requestService{validator: validator.New()}
}

func (s *requestService) ValidateRequest(ctx echo.Context, request interface{}) (interface{}, error) {
	if err := ctx.Bind(request); err != nil {
		return nil, err
	}

	if err := s.validator.Struct(request); err != nil {
		return nil, errors.New(err.Error())
	}

	return request, nil
}