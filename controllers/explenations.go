package controllers

import (
	"encoding/json"
	"net/http"

	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/models"
	govalidator "gopkg.in/asaskevich/govalidator.v4"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

//ExplenationsCtrl /explenations
type ExplenationsCtrl struct{}

func (explenationsCtrl ExplenationsCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "ExplenationsCtrl",
	})
}

//List all explenations
func (explenationsCtrl ExplenationsCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	explenationID := vars["id"]

	var explenation models.Explenation

	if explenationID != "" {
		if err := explenation.FindByID(explenationID); err != nil {
			jsonErr := helpers.GenerateJSONError(helpers.ErrRecordNotFound, req.Header)
			r.JSON(res, jsonErr.StatusCode, jsonErr)
			return
		}
		r.JSON(res, 200, explenation)
		return
	}

	explenations, err := explenation.FindAll()
	if err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrInternalServerError, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.JSON(res, 200, explenations)
	return
}

//Create property
func (explenationsCtrl ExplenationsCtrl) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var explenation models.Explenation
	if err := decoder.Decode(&explenation); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	// validate login request
	if _, err := govalidator.ValidateStruct(explenation); err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrUnprocessableEntity, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	if err := explenation.Save(); err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrInternalServerError, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.JSON(res, 200, explenation)
	return
}

//Update explenation
func (explenationsCtrl ExplenationsCtrl) Update(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	explenationID := vars["id"]

	var explenation models.Explenation
	if err := explenation.FindByID(explenationID); err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrInternalServerError, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var newExplenation models.Explenation
	if err := decoder.Decode(&newExplenation); err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrBadRequest, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	// validate update in request
	_, err := govalidator.ValidateStruct(newExplenation)
	if err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrUnprocessableEntity, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	newExplenation.ID = explenation.ID
	newExplenation.CreatedAt = explenation.CreatedAt

	err = newExplenation.Save()
	if err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrInternalServerError, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.JSON(res, 200, newExplenation)
	return
}

//Delete explenation
func (explenationsCtrl ExplenationsCtrl) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	explenationID := vars["id"]

	var explenation models.Explenation
	if err := explenation.FindByID(explenationID); err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrRecordNotFound, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	if err := explenation.Delete(); err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrInternalServerError, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.Text(res, 204, "")
	return
}
