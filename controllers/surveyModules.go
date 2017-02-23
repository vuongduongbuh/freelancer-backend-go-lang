package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"gopkg.in/mgo.v2/bson"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//SurveyModulesCtrl is the controller for /modules
type SurveyModulesCtrl struct{}

func (surveyModulesCtrl SurveyModulesCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "SurveyModulesCtrl",
	})
}

//List a/all modules
func (surveyModulesCtrl SurveyModulesCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	id := vars["id"]
	showDrafts := req.FormValue("showDrafts")

	var surveyModule models.SurveyModule

	if id != "" {
		surveyModule.FindByID(id)

		if surveyModule.ID == "" {
			r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
			return
		}

		r.JSON(res, 200, surveyModule)
		return
	}

	surveyModules, err := surveyModule.FindAll()
	if showDrafts == "true" {
		surveyModules, err = surveyModule.FindAllIncludingDrafts()
	}

	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyModules)
	return
}

//Create a new module
func (surveyModulesCtrl SurveyModulesCtrl) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var surveyModule models.SurveyModule
	if err := decoder.Decode(&surveyModule); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	surveyModule.IsDraft = true

	if err := surveyModule.Save(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyModule)
	return
}

//Update a module
func (surveyModulesCtrl SurveyModulesCtrl) Update(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	surveyModulesID := vars["id"]

	var surveyModule models.SurveyModule
	if err := surveyModule.FindByID(surveyModulesID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var newSurveyModule models.SurveyModule
	if err := decoder.Decode(&newSurveyModule); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	newSurveyModule.ID = surveyModule.ID
	newSurveyModule.CreatedAt = surveyModule.CreatedAt

	err := newSurveyModule.Save()
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, newSurveyModule)
}

//Delete survey module
func (surveyModulesCtrl SurveyModulesCtrl) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	id := vars["id"]

	var surveyModule models.SurveyModule
	if err := surveyModule.FindByID(id); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	if err := surveyModule.Delete(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}

//Publish changes the modules draft status to false
func (surveyModulesCtrl SurveyModulesCtrl) Publish(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	surveyModulesID := vars["id"]

	var surveyModule models.SurveyModule
	if err := surveyModule.FindByID(surveyModulesID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	if err := surveyModule.SetDraftStatus(false); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}

//Unpublish changes the modules draft status to true
func (surveyModulesCtrl SurveyModulesCtrl) Unpublish(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	surveyModulesID := vars["id"]

	var surveyModule models.SurveyModule
	if err := surveyModule.FindByID(surveyModulesID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	if err := surveyModule.SetDraftStatus(true); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}

//AddQuestion validates and adds a single question to the given module
func (surveyModulesCtrl SurveyModulesCtrl) AddQuestion(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	moduleID := vars["id"]

	if isObjectID := bson.IsObjectIdHex(moduleID); !isObjectID {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}

	var surveyModule models.SurveyModule
	surveyModule.ID = bson.ObjectIdHex(moduleID)

	decoder := json.NewDecoder(req.Body)
	var question models.Question

	estimatedQuestionDuration := float32(12000)
	if estimatedDuration, err := strconv.ParseFloat(os.Getenv("ESTIMATED_QUESTION_DURATION"), 2); err == nil {
		estimatedQuestionDuration = float32(estimatedDuration)
	}

	question.EstimatedDuration = float32(estimatedQuestionDuration)
	if err := decoder.Decode(&question); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	createdQuestion, err := surveyModule.AddQuestion(question)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 201, createdQuestion)
}

//GetQuestion gets a single question from the given module
func (surveyModulesCtrl SurveyModulesCtrl) GetQuestion(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	moduleID := vars["id"]
	questionID := vars["questionID"]

	if isObjectID := bson.IsObjectIdHex(questionID); !isObjectID {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}
	if isObjectID := bson.IsObjectIdHex(moduleID); !isObjectID {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}

	var surveyModule models.SurveyModule
	surveyModule.ID = bson.ObjectIdHex(moduleID)

	question, err := surveyModule.FindQuestionByObjectID(bson.ObjectIdHex(questionID))
	if err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, question)
}

//UpdateQuestion updates a single question from the given module
func (surveyModulesCtrl SurveyModulesCtrl) UpdateQuestion(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	moduleID := vars["id"]
	questionID := vars["questionID"]

	decoder := json.NewDecoder(req.Body)
	var updatedQuestion models.Question
	if err := decoder.Decode(&updatedQuestion); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	if isObjectID := bson.IsObjectIdHex(questionID); !isObjectID {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}

	updatedQuestion.ID = bson.ObjectIdHex(questionID)
	if err := updatedQuestion.Validate(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if isObjectID := bson.IsObjectIdHex(moduleID); !isObjectID {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}

	var surveyModule models.SurveyModule
	surveyModule.ID = bson.ObjectIdHex(moduleID)

	resQuestion, err := surveyModule.UpdateQuestion(updatedQuestion)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, resQuestion)
}

//DeleteQuestion removes a single question from the given module
func (surveyModulesCtrl SurveyModulesCtrl) DeleteQuestion(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	moduleID := vars["id"]
	questionID := vars["questionID"]

	if isObjectID := bson.IsObjectIdHex(questionID); !isObjectID {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}
	if isObjectID := bson.IsObjectIdHex(moduleID); !isObjectID {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}

	var surveyModule models.SurveyModule
	surveyModule.ID = bson.ObjectIdHex(moduleID)

	if err := surveyModule.DeleteQuestion(bson.ObjectIdHex(questionID)); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}

//SortQuestions sorts the questions and return result
func (surveyModulesCtrl SurveyModulesCtrl) SortQuestions(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	moduleID := vars["id"]

	if isObjectID := bson.IsObjectIdHex(moduleID); !isObjectID {
		r.JSON(res, 422, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var questionIDs []bson.ObjectId
	if err := decoder.Decode(&questionIDs); err != nil {
		if err != io.EOF {
			r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrEmptyRequestBody.Error(), req.Header))
		return
	}

	var surveyModule models.SurveyModule
	if err := surveyModule.FindByID(moduleID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	if len(surveyModule.Questions) != len(questionIDs) {
		r.JSON(res, 422, helpers.GenerateErrorResponse(models.ErrQuestionsCountMissmatch.Error(), req.Header))
		return
	}

	questionSortingMap := make(map[bson.ObjectId]int)
	for sortIndex, questionID := range questionIDs {
		questionSortingMap[questionID] = sortIndex
	}

	for questionIndex, question := range surveyModule.Questions {
		surveyModule.Questions[questionIndex].SortIndex = questionSortingMap[question.ID]
	}

	if err := surveyModule.Save(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, surveyModule)
}
