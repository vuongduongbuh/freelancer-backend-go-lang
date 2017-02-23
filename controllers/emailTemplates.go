package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/models"
	"gopkg.in/asaskevich/govalidator.v4"
)

//EmailTemplatesCtrl is the controller for /emailTemplate
type EmailTemplatesCtrl struct{}

func (emailTemplatesCtrl EmailTemplatesCtrl) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"controller": "EmailTemplatesCtrl",
	})
}

//List all email templates
func (emailTemplatesCtrl EmailTemplatesCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	id := vars["id"]

	if id != "" {
		var email models.EMailTemplate
		if err := email.FindByID(id); err != nil {
			r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
			return
		}

		r.JSON(res, 200, email)
		return
	}

	var email models.EMailTemplate
	emails, err := email.FindAll()
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, emails)
	return
}

//Create email template
func (emailTemplatesCtrl EmailTemplatesCtrl) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var emailTemplate models.EMailTemplate
	if err := decoder.Decode(&emailTemplate); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	// validate email template
	_, err := govalidator.ValidateStruct(emailTemplate)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if err := emailTemplate.Save(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, emailTemplate)
}

//Update email template
func (emailTemplatesCtrl EmailTemplatesCtrl) Update(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	emailID := vars["id"]

	var eMailTemplate models.EMailTemplate
	if err := eMailTemplate.FindByID(emailID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var updatedEMailTemplate models.EMailTemplate
	if err := decoder.Decode(&updatedEMailTemplate); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	// validate email template request
	_, err := govalidator.ValidateStruct(updatedEMailTemplate)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	updatedEMailTemplate.ID = eMailTemplate.ID
	updatedEMailTemplate.CreatedAt = eMailTemplate.CreatedAt

	if err := updatedEMailTemplate.Save(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, updatedEMailTemplate)
}

//Delete email template
func (emailTemplatesCtrl EmailTemplatesCtrl) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	emailID := vars["id"]

	var eMailTemplate models.EMailTemplate
	if err := eMailTemplate.FindByID(emailID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	if err := eMailTemplate.Delete(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}
