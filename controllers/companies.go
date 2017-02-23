package controllers

import (
	"encoding/json"
	"net/http"

	"gitlab.com/slsurvey/slsurvey-srv/helpers"

	"io"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/models"
	"gopkg.in/asaskevich/govalidator.v4"
)

//CompaniesCtrl is the controller for /companies
type CompaniesCtrl struct{}

func (companiesCtrl CompaniesCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "CompaniesCtrl",
	})
}

//List all companies
func (companiesCtrl CompaniesCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	id := vars["id"]

	var company models.Company

	if id != "" {
		company.FindByID(id)

		if company.ID == "" {
			r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
			return
		}

		r.JSON(res, 200, company)
		return
	}

	companies, err := company.FindAll()

	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, companies)
	return
}

//Create property
func (companiesCtrl CompaniesCtrl) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var company models.Company
	if err := decoder.Decode(&company); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	// validate login request
	if _, err := govalidator.ValidateStruct(company); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if err := company.Save(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, company)
}

//Update company
func (companiesCtrl CompaniesCtrl) Update(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	companyID := vars["id"]

	var company models.Company
	if err := company.FindByID(companyID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var newCompany models.Company
	if err := decoder.Decode(&newCompany); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	// validate login request
	_, err := govalidator.ValidateStruct(newCompany)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	newCompany.ID = company.ID
	newCompany.CreatedAt = company.CreatedAt

	err = newCompany.Save()
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, newCompany)
}

//Delete company
func (companiesCtrl CompaniesCtrl) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	companyID := vars["id"]

	var company models.Company
	if err := company.FindByID(companyID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	if err := company.Delete(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}
