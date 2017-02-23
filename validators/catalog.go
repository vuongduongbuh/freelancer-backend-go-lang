package validators

import (
	"errors"

	"github.com/Sirupsen/logrus"

	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/models"
	"gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2/bson"
)

func registerSurveyCatalogValidators() {
	registerOptionalIfDraftValidator()
	registerExistingCompanyValidator()
}

func surveyCast(v interface{}) (models.Survey, error) {
	survey, isCorrectType := v.(models.Survey)
	if !isCorrectType {
		return models.Survey{}, errors.New("can't cast to Survey")
	}
	return survey, nil
}

func registerOptionalIfDraftValidator() {
	govalidator.CustomTypeTagMap.Set("optionalIfDraft", func(i interface{}, o interface{}) bool {
		if survey, err := surveyCast(i); err == nil {
			if err := survey.Validate(); err != nil {
				helpers.GetLogger().WithFields(logrus.Fields{
					"validator": "OptionalIfDraft",
				}).Info(err)
				return false
			}
			return true
		}
		return false
	})
}

func objectIDCast(v interface{}) (bson.ObjectId, error) {
	objectID, isCorrectType := v.(bson.ObjectId)
	if !isCorrectType {
		return bson.NewObjectId(), errors.New("can't cast to bson.ObjectId")
	}
	return objectID, nil
}

func registerExistingCompanyValidator() {
	govalidator.CustomTypeTagMap.Set("existingCompany", func(i interface{}, o interface{}) bool {
		if companyID, err := objectIDCast(i); err == nil {
			isExistingCompanyErr := models.EnsureDocumentExistence(models.CollectionCompanyProperty, companyID)
			return isExistingCompanyErr == nil
		}
		return false
	})
}
