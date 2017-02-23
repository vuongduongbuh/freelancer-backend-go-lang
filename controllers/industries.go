package controllers

import (
	"net/http"

	"gitlab.com/slsurvey/slsurvey-srv/helpers"

	"github.com/Sirupsen/logrus"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//IndustriesCtrl is the controller for /industries
type IndustriesCtrl struct{}

func (industriesCtrl IndustriesCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "IndustriesCtrl",
	})
}

func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

//List all companies
func (industriesCtrl IndustriesCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	var company models.Company
	companies, err := company.FindAll()
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	industries := make([]string, 0)
	for _, company := range companies {
		industries = appendIfMissing(industries, company.Industry)
	}

	r.JSON(res, 200, industries)
}
