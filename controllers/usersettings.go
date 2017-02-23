package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/middlewares"
	"gitlab.com/slsurvey/slsurvey-srv/models"

	"github.com/Sirupsen/logrus"
	"github.com/unrolled/render"
	"gopkg.in/asaskevich/govalidator.v4"
)

//UserSettingsCtrl /usersettings
type UserSettingsCtrl struct{}

func (userSettingsCtrl UserSettingsCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "UserSettingsCtrl",
	})
}

type userSettingsRequestModel struct {
	Token    string `json:"token"`
	Type     string `json:"type"`
	Language string `json:"language"`
}

//Save current userSettings token to ticket
func (userSettingsCtrl UserSettingsCtrl) Save(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	surveyTicket, err := middlewares.GetSurveyTicketFromContext(req)
	if err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var userSettingsRequestData userSettingsRequestModel
	if err := decoder.Decode(&userSettingsRequestData); err != nil {
		if err != io.EOF {
			r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 422, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return

	}

	if _, err := govalidator.ValidateStruct(&userSettingsRequestData); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if err := surveyTicket.SetUserSettings(userSettingsRequestData.Token, userSettingsRequestData.Type, userSettingsRequestData.Language); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}

//Delete current userSettings token from ticket
func (userSettingsCtrl UserSettingsCtrl) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	surveyTicket, err := middlewares.GetSurveyTicketFromContext(req)
	if err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	if err := surveyTicket.DeletePushToken(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}
