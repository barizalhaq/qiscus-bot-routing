package controllers

import (
	"bot-routing-engine/entities/viewmodel"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type uploadController struct{}

func NewUploadController() *uploadController {
	return &uploadController{}
}

func (controller *uploadController) Upload(ctx echo.Context) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, viewmodel.ErrorResponse{Message: err.Error()})
		return err
	}

	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, viewmodel.ErrorResponse{Message: err.Error()})
		return err
	}
	defer src.Close()

	filename := file.Filename
	if os.Getenv("ALL_IN_ONE_JSON_ROUTE") == "true" {
		filename = fmt.Sprintf("%s%s", "layer", filepath.Ext(filename))
	}

	dir := fmt.Sprintf("layer/%s", filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll("./layer", 0700)
	}

	dst, err := os.Create(dir)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, viewmodel.ErrorResponse{Message: err.Error()})
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		ctx.JSON(http.StatusBadRequest, viewmodel.ErrorResponse{Message: err.Error()})
		return err
	}

	ctx.JSON(http.StatusOK, "success")
	return nil
}
