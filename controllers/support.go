package controllers

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//SupportCtrl is the controller for /support
type SupportCtrl struct{}

// Return light version list survey ticket 
func (supportCtrl SupportCtrl) LightTicketsList(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	id := vars["id"]

	if id == "" {
		jsonErr := helpers.GenerateJSONError(helpers.ErrBadRequest, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	var surveyCatalog models.SurveyCatalog

	if err := surveyCatalog.FindByID(id); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	if(!surveyCatalog.IsActive || surveyCatalog.IsDone) {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	surveyTicket,err := surveyCatalog.GetSupportTickets()
	if err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.JSON(res, 200, surveyTicket)
	return
}