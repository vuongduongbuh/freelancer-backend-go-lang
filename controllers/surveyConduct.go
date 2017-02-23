package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/middlewares"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//ConductCtrl is the controller for /conduct
type ConductCtrl struct{}

func (conductCtrl ConductCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "ConductCtrl",
	})
}

//GetAnswers get answers of already conducted questions
func (conductCtrl ConductCtrl) GetAnswers(res http.ResponseWriter, req *http.Request) {
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

func parseToFloat(floatInterface interface{}) float32 {
	result := float32(1200)
	switch i := floatInterface.(type) {
	case float32:
		result = i
	default:
		result = float32(1200)
	}

	if result < 10 {
		result = float32(1200)
	}
	return result
}

//Conduct a given survey
func (conductCtrl ConductCtrl) Conduct(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	todoID := vars["id"]

	surveyTicket, err := middlewares.GetSurveyTicketFromContext(req)
	if err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var questionAnswersQuickFix []models.SurveyTicketConductRequestModelQuickFix
	if err := decoder.Decode(&questionAnswersQuickFix); err != nil {
		if err != io.EOF {
			r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	//TODO: remove quick fix with v2 route
	questionAnswers := make([]models.SurveyTicketConductRequestModel, 0)
	for _, questionAnswerQuickFix := range questionAnswersQuickFix {
		questionAnswer := models.SurveyTicketConductRequestModel{
			Answer:           questionAnswerQuickFix.Answer,
			QuestionID:       questionAnswerQuickFix.QuestionID,
			MeasuredDuration: parseToFloat(questionAnswerQuickFix.MeasuredDuration),
		}
		questionAnswers = append(questionAnswers, questionAnswer)
	}

	if err := surveyTicket.AnswerQuestions(todoID, questionAnswers, ""); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}

//Submit a given survey
func (conductCtrl ConductCtrl) Submit(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	todoID := vars["id"]

	surveyTicket, err := middlewares.GetSurveyTicketFromContext(req)
	if err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var questionAnswersQuickFix []models.SurveyTicketConductRequestModelQuickFix
	if err := decoder.Decode(&questionAnswersQuickFix); err != nil {
		if err != io.EOF {
			r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	//TODO: remove quick fix with v2 route
	questionAnswers := make([]models.SurveyTicketConductRequestModel, 0)
	for _, questionAnswerQuickFix := range questionAnswersQuickFix {
		questionAnswer := models.SurveyTicketConductRequestModel{
			Answer:           questionAnswerQuickFix.Answer,
			QuestionID:       questionAnswerQuickFix.QuestionID,
			MeasuredDuration: parseToFloat(questionAnswerQuickFix.MeasuredDuration),
		}
		questionAnswers = append(questionAnswers, questionAnswer)
	}

	if err == nil {
		if err := surveyTicket.AnswerQuestions(todoID, questionAnswers, ""); err != nil {
			r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
	}

	if err := surveyTicket.SubmitTodo(todoID, ""); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyTicket)
}
