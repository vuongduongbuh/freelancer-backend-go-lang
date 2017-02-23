package controllers

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//AssetsCtrl handels all /assets requests
type AssetsCtrl struct{}

func (asssetsCtrl AssetsCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "AssetsCtrl",
	})
}

//UploadImages checks the file and saves it to ./assets/images
func (asssetsCtrl AssetsCtrl) UploadImages(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	req.ParseMultipartForm(0)

	file, handler, err := req.FormFile("file")
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}
	defer file.Close()

	var asset models.Asset
	if err := asset.Create(models.AssetTypeImage, handler, file); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if err := asset.Upload(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 201, asset)
	return
}
