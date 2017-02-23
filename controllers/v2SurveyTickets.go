package controllers

import (
	"net/http"

	"gopkg.in/mgo.v2/bson"

	"github.com/Sirupsen/logrus"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/middlewares"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//TicketsCtrlV2 is the controller for /api/v2/ticket
type TicketsCtrlV2 struct{}

func (ticketsCtrl TicketsCtrlV2) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "TicketsCtrl - V2",
	})
}

//List returns and enriches the survey ticket
func (ticketsCtrl TicketsCtrlV2) List(res http.ResponseWriter, req *http.Request) {
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

	todoSize := len(surveyTicket.Todos)
	if todoSize <= 0 {
		r.JSON(res, 200, surveyTicket)
		return
	}

	//TODO: remove hack
	questionToChange := bson.ObjectIdHex("58297ab0258b2703360b0369")
	lastTodoIdx := todoSize - 1
	for questionIdx, question := range surveyTicket.Todos[lastTodoIdx].Questions {
		if question.ID != questionToChange {
			continue
		}
		question.Keyboard.IsOptional = true
		surveyTicket.Todos[lastTodoIdx].Questions[questionIdx] = question
	}

	r.JSON(res, 200, surveyTicket)
	return
}
