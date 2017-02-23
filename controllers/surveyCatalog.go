package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	govalidator "gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2/bson"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//SurveyCatalogCtrl is the controller for /catalogs
type SurveyCatalogCtrl struct{}

func (surveyCatalogCtrl SurveyCatalogCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "SurveyCatalogCtrl",
	})
}

//List a/all catalogs
func (surveyCatalogCtrl SurveyCatalogCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	id := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog

	if id != "" {
		surveyCatalog.FindByID(id)
		if !surveyCatalog.ID.Valid() {
			r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
			return
		}

		r.JSON(res, 200, surveyCatalog)
		return
	}

	surveyCatalogs, err := surveyCatalog.FindAll()
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyCatalogs)
	return
}

type initalCatalogCreationRequestData struct {
	Name      string        `json:"name" valid:"required"`
	CompanyID bson.ObjectId `json:"companyId" valid:"existingCompany,required"`
}

//Create a new module
func (surveyCatalogCtrl SurveyCatalogCtrl) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var requestData initalCatalogCreationRequestData
	if err := decoder.Decode(&requestData); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	//validate the POST request data
	_, err := govalidator.ValidateStruct(requestData)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var company models.Company
	if err := company.FindByObjectID(requestData.CompanyID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	surveyCatalog := models.SurveyCatalog{
		Name:    requestData.Name,
		Company: company,
		IsDraft: true,
	}

	if err := surveyCatalog.Save(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 201, surveyCatalog)
	return
}

//GetBaseSurveyModules get selected base survey modules
func (surveyCatalogCtrl SurveyCatalogCtrl) GetBaseSurveyModules(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog
	if err := surveyCatalog.FindByID(catalogID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	moduleIDs, err := surveyCatalog.GetSurveyModulesIDs(catalogID, 0)
	if err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, moduleIDs)
	return
}

//AddBaseSurveyModules add base survey to a draft catalog
func (surveyCatalogCtrl SurveyCatalogCtrl) AddBaseSurveyModules(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog

	if err := surveyCatalog.ValidateAndAddSurveyModules(catalogID, 0, req.Body); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyCatalog)
	return
}

//GetRecurringSurveyModules get selected recurring survey modules
func (surveyCatalogCtrl SurveyCatalogCtrl) GetRecurringSurveyModules(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog
	if err := surveyCatalog.FindByID(catalogID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	moduleIDs, err := surveyCatalog.GetSurveyModulesIDs(catalogID, 1)
	if err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, moduleIDs)
	return
}

//AddRecurringSurveyModules add recurring to a draft catalog
func (surveyCatalogCtrl SurveyCatalogCtrl) AddRecurringSurveyModules(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog

	if err := surveyCatalog.ValidateAndAddSurveyModules(catalogID, 1, req.Body); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyCatalog)
	return
}

//GetEndSurveyModules get selected end survey modules
func (surveyCatalogCtrl SurveyCatalogCtrl) GetEndSurveyModules(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog
	if err := surveyCatalog.FindByID(catalogID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	moduleIDs, err := surveyCatalog.GetSurveyModulesIDs(catalogID, 2)
	if err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, moduleIDs)
	return
}

//AddEndSurveyModules adds the finishing survey to a draft catalog
func (surveyCatalogCtrl SurveyCatalogCtrl) AddEndSurveyModules(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog

	if err := surveyCatalog.ValidateAndAddSurveyModules(catalogID, 2, req.Body); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyCatalog)
	return
}

//GetSettings adds the settings to the catalog
func (surveyCatalogCtrl SurveyCatalogCtrl) GetSettings(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog
	if err := surveyCatalog.FindByID(catalogID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyCatalog.RawCatalogSettings)
	return
}

//AddSettings adds the settings to the catalog
func (surveyCatalogCtrl SurveyCatalogCtrl) AddSettings(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog
	if err := surveyCatalog.FindByID(catalogID); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var requestData models.SurveyCatalogAddSettingsRequestModel
	if err := decoder.Decode(&requestData); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	if err := surveyCatalog.AddSurveySettings(requestData); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyCatalog)
	return
}

type setEMailTemplateRequestModel struct {
	EMailTemplateID bson.ObjectId `json:"eMailTemplateId" valid:"required"`
}

//SetEMailTemplateID adds the emailTemplateObjectID
func (surveyCatalogCtrl SurveyCatalogCtrl) SetEMailTemplateID(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog
	if err := surveyCatalog.FindByID(catalogID); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var requestData setEMailTemplateRequestModel
	if err := decoder.Decode(&requestData); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	if !requestData.EMailTemplateID.Valid() {
		r.JSON(res, 422, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}

	var emailTemplate models.EMailTemplate
	if err := emailTemplate.FindByObjectID(requestData.EMailTemplateID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	surveyCatalog.EMailTemplateID = requestData.EMailTemplateID

	if err := surveyCatalog.Save(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
	return
}

//Publish creates the tickets for the given survey
func (surveyCatalogCtrl SurveyCatalogCtrl) Publish(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	var surveyCatalog models.SurveyCatalog
	if err := surveyCatalog.FindByID(catalogID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if err := surveyCatalog.Publish(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyCatalog)
	return
}

//Delete survey module
func (surveyCatalogCtrl SurveyCatalogCtrl) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	catalogID := vars["catalogId"]

	if catalogID == "" {
		jsonErr := helpers.GenerateJSONError(helpers.ErrBadRequest, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	var surveyCatalog models.SurveyCatalog
	if err := surveyCatalog.FindByID(catalogID); err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrRecordNotFound, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	if err := surveyCatalog.Delete(); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.Text(res, 204, "")
	return
}
