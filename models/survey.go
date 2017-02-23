package models

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	govalidator "gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2/bson"
)

var (
	//ErrModuleNotFound survey module not found
	ErrModuleNotFound = errors.New("module_not_found")
	//ErrModuleIsDraft survey module found but isDraft = true
	ErrModuleIsDraft = errors.New("module_is_draft")
)

//Survey model
type Survey struct {
	BaseModel `bson:",inline"`

	Modules []SurveyModule `json:"modules,omitempty" bson:",omitempty" valid:"required"`
}

//getLogger
func (survey *Survey) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "Survey",
	})
}

//Validate the given Survey
func (survey *Survey) Validate() error {
	_, err := govalidator.ValidateStruct(survey)
	if err != nil {
		return err
	}

	for _, module := range survey.Modules {
		if err := module.Validate(); err != nil {
			return err
		}
	}

	return nil
}

//Create the given Survey
func (survey *Survey) Create(moduleIDs []bson.ObjectId) error {
	survey.SetID()
	survey.UpdatedAt = time.Now()

	surveyModules := make([]SurveyModule, 0)
	for _, moduleID := range moduleIDs {
		var surveyModule SurveyModule
		if err := surveyModule.FindByObjectID(moduleID); err != nil {
			survey.getLogger().WithFields(logrus.Fields{
				"function": "Create",
			}).Info(ErrModuleNotFound)
			return ErrModuleNotFound
		}
		if surveyModule.IsDraft {
			survey.getLogger().WithFields(logrus.Fields{
				"function": "Create",
			}).Info(ErrModuleIsDraft)
			return ErrModuleIsDraft
		}
		surveyModules = append(surveyModules, surveyModule)
	}

	survey.Modules = surveyModules

	return nil
}
