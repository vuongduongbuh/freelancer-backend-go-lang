package controllers

import (
	"encoding/json"
	"net/http"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/middlewares"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//ConductCtrlV2 is the controller for /conduct
type ConductCtrlV2 struct{}

func (conductCtrl ConductCtrlV2) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "ConductCtrl - V2",
	})
}

//GetAnswers get answers of already conducted questions
func (conductCtrl ConductCtrlV2) GetAnswers(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	todoID := vars["id"]

	surveyTicket, err := middlewares.GetSurveyTicketFromContext(req)
	if err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	answeredQuestions, err := surveyTicket.GetAnsweredQuestions(todoID)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, answeredQuestions)
}

//Conduct a given survey
func (conductCtrl ConductCtrlV2) Conduct(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	todoID := vars["id"]

	surveyTicket, err := middlewares.GetSurveyTicketFromContext(req)
	if err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var questionAnswers []models.SurveyTicketConductRequestModel
	if err := decoder.Decode(&questionAnswers); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	clientIdentifier := req.Header.Get(os.Getenv("APPLICATION_IDENTIFIER_HEADER"))
	if err := surveyTicket.AnswerQuestions(todoID, questionAnswers, clientIdentifier); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
	return
}

//Submit a given survey
func (conductCtrl ConductCtrlV2) Submit(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	defer req.Body.Close()

	vars := mux.Vars(req)
	todoID := vars["id"]

	surveyTicket, err := middlewares.GetSurveyTicketFromContext(req)
	if err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var questionAnswers []models.SurveyTicketConductRequestModel
	if err := decoder.Decode(&questionAnswers); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	applicationIdentifier := req.Header.Get(os.Getenv("APPLICATION_IDENTIFIER_HEADER"))

	if err == nil {
		if err := surveyTicket.AnswerQuestions(todoID, questionAnswers, applicationIdentifier); err != nil {
			r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
	}

	if err := surveyTicket.SubmitTodo(todoID, applicationIdentifier); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyTicket)
	return
}
