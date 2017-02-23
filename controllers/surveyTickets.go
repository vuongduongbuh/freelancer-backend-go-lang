package controllers

import (
	"net/http"
	"github.com/Sirupsen/logrus"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/middlewares"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//TicketsCtrl is the controller for /todos
type TicketsCtrl struct{}

func (ticketsCtrl TicketsCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "TicketsCtrl",
	})
}

//List returns a users ticket if he has a valid token
//it is a depcreated endpoint that will be released removed in a future release
func (ticketsCtrl TicketsCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	surveyTicket, err := middlewares.GetSurveyTicketFromContext(req)
	if err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	if err := surveyTicket.EnrichActiveTodos(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	//TODO: remove when fixed
	for todoIdx, todo := range surveyTicket.Todos {
		for questionIdx, question := range todo.Questions {
			if question.Keyboard.Type != models.QuestionKeyboardTypeInformation {
				continue
			}
			surveyTicket.Todos[todoIdx].Questions[questionIdx].Answer = "quickfix"
		}
	}

	r.JSON(res, 200, surveyTicket)
	return
}

//ListV2 returns a users ticket if he has a valid token
func (ticketsCtrl TicketsCtrl) ListV2(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	surveyTicket, err := middlewares.GetSurveyTicketFromContext(req)
	if err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	if err := surveyTicket.EnrichActiveTodos(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyTicket)
	return
}
