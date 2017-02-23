package models

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//CollectionSurveyModuleProperty mongodb store name
var CollectionSurveyModuleProperty = "surveyModule"

//SurveyModule model
type SurveyModule struct {
	BaseModel `bson:",inline"`

	Name              string     `json:"name" valid:"required"`
	Questions         []Question `json:"questions" bson:",omitempty"`
	EstimatedDuration float32    `json:"estimatedDuration"`

	IsDraft bool `json:"isDraft"`
}

func (surveyModule SurveyModule) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model":      "SurveyModule",
		"collection": CollectionSurveyModuleProperty,
	})
}

//Validate the given questions
func (surveyModule *SurveyModule) Validate() error {
	_, err := govalidator.ValidateStruct(surveyModule)
	if err != nil {
		return err
	}

	for _, question := range surveyModule.Questions {
		err := question.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

//Save or updates the given object
func (surveyModule *SurveyModule) Save() error {
	if err := surveyModule.Validate(); err != nil {
		return err
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	SurveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	surveyModule.SetID()
	surveyModule.UpdatedAt = time.Now()

	for questionIdx, question := range surveyModule.Questions {
		question.SetID()
		question.UpdatedAt = time.Now()

		surveyModule.Questions[questionIdx] = question
	}

	_, err := SurveyModuleCol.UpsertId(surveyModule.ID, bson.M{"$set": surveyModule})
	if err != nil {
		return err
	}
	return surveyModule.UpdateQuestionSort()
}

//FindAll finds all modules
func (surveyModule *SurveyModule) FindAll() ([]SurveyModule, error) {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return nil, dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	surveyModules := make([]SurveyModule, 0)
	err := surveyModuleCol.Find(bson.M{"isdraft": false}).All(&surveyModules)
	if err != nil {
		surveyModule.getLogger().WithFields(logrus.Fields{
			"query": "Find(nil)",
		}).Info(err)
		return nil, ErrRecordsFetch
	}

	return surveyModules, nil
}

//FindAllIncludingDrafts finds all modules
func (surveyModule *SurveyModule) FindAllIncludingDrafts() ([]SurveyModule, error) {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return nil, dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	surveyModules := make([]SurveyModule, 0)
	err := surveyModuleCol.Find(nil).All(&surveyModules)
	if err != nil {
		surveyModule.getLogger().WithFields(logrus.Fields{
			"query": "Find(nil)",
		}).Info(err)
		return nil, ErrRecordsFetch
	}

	return surveyModules, nil
}

//FindByID finds a module
func (surveyModule *SurveyModule) FindByID(id string) error {
	if isObjectID := bson.IsObjectIdHex(id); !isObjectID {
		return ErrInvalidObjectID
	}

	return surveyModule.FindByObjectID(bson.ObjectIdHex(id))
}

//FindByObjectID finds a module
func (surveyModule *SurveyModule) FindByObjectID(surveyModuleID bson.ObjectId) error {
	if isObjectID := surveyModuleID.Valid(); !isObjectID {
		return ErrInvalidObjectID
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	if err := surveyModuleCol.FindId(surveyModuleID).One(&surveyModule); err != nil {
		surveyModule.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("FindId(%s))", surveyModuleID),
			"error": err.Error(),
		}).Info(err)
		return ErrRecordNotRetrievable
	}

	return nil
}

//Delete module
func (surveyModule *SurveyModule) Delete() error {
	if !surveyModule.ID.Valid() {
		return ErrRecordNotFound
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	err := surveyModuleCol.RemoveId(surveyModule.ID)
	if err != nil {
		surveyModule.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("RemoveId(%s))", surveyModule.ID),
			"error": err.Error(),
		}).Error(err)
		return ErrRecordNotDeleteable
	}

	return nil
}

//SetDraftStatus module
func (surveyModule *SurveyModule) SetDraftStatus(isDraft bool) error {
	if !surveyModule.ID.Valid() {
		return ErrRecordNotFound
	}

	updateDraftStatusChange := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"isdraft": isDraft,
			},
		},
		ReturnNew: false,
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	_, err := surveyModuleCol.Find(bson.M{"_id": surveyModule.ID}).Apply(updateDraftStatusChange, &surveyModule)
	if err != nil {
		surveyModule.getLogger().Error(err)
		return ErrRecordNotDeleteable
	}
	return nil
}

//UpdateEstimatedDuration sums the value up in the database to minimize RAM consumption
func (surveyModule *SurveyModule) UpdateEstimatedDuration() error {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	matchSurveyModuleIDOperation := bson.M{
		"$match": bson.M{"_id": surveyModule.ID},
	}
	unwindSurveyModuleQuestionsOperation := bson.M{
		"$unwind": "$questions",
	}
	sumQuestionsEstimatedDurationOperation := bson.M{
		"$group": bson.M{
			"_id": nil,
			"estimatedDuration": bson.M{
				"$sum": "$questions.estimatedduration",
			},
		},
	}

	operations := []bson.M{matchSurveyModuleIDOperation, unwindSurveyModuleQuestionsOperation, sumQuestionsEstimatedDurationOperation}
	aggregationPipe := surveyModuleCol.Pipe(operations)

	aggregationResult := bson.M{}
	if err := aggregationPipe.One(&aggregationResult); err != nil {
		if err != mgo.ErrNotFound {
			return err
		}
	}

	updateEstimatedDurationChange := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"estimatedduration": aggregationResult["estimatedDuration"],
			},
		},
		ReturnNew: true,
	}

	_, err := surveyModuleCol.Find(bson.M{"_id": surveyModule.ID}).Apply(updateEstimatedDurationChange, &surveyModule)
	if err != nil {
		surveyModule.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("UpdateId(%s))", surveyModule.ID),
			"error": err.Error(),
		}).Error(err)
		return ErrRecordNotDeleteable
	}
	return nil
}

//UpdateQuestionSort sorts the questions by sort indey
func (surveyModule *SurveyModule) UpdateQuestionSort() error {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	updateSortOrderChange := mgo.Change{
		Update: bson.M{
			"$push": bson.M{
				"questions": bson.M{
					"$each":  []bson.M{},
					"$sort":  bson.M{"sortindex": 1},
					"$slice": -len(surveyModule.Questions),
				},
			},
		},
		ReturnNew: true,
	}

	if _, err := surveyModuleCol.Find(bson.M{"_id": surveyModule.ID}).Apply(updateSortOrderChange, &surveyModule); err != nil {
		surveyModule.getLogger().Error(err)
		return ErrRecordNotDeleteable
	}

	return nil
}

//FindQuestionByObjectID in the current SurveyModule
func (surveyModule *SurveyModule) FindQuestionByObjectID(questionID bson.ObjectId) (Question, error) {
	if !surveyModule.ID.Valid() {
		return Question{}, ErrRecordNotFound
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return Question{}, dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	matchQuestionIDOperation := bson.M{
		"$match": bson.M{"questions._id": questionID},
	}

	filterRelevantQuestionsOperation := bson.M{
		"$project": bson.M{
			"questions": bson.M{
				"$filter": bson.M{
					"input": "$questions",
					"as":    "questions",
					"cond": bson.M{
						"$eq": []interface{}{
							"$$questions._id",
							questionID,
						},
					},
				},
			},
		},
	}

	operations := []bson.M{matchQuestionIDOperation, filterRelevantQuestionsOperation}
	pipe := surveyModuleCol.Pipe(operations)

	var queryResult SurveyModule
	if err := pipe.One(&queryResult); err != nil {
		return Question{}, err
	}

	if len(queryResult.Questions) < 1 {
		return Question{}, ErrRecordNotFound
	}

	return queryResult.Questions[0], nil
}

//AddQuestion to SurveyModule
func (surveyModule *SurveyModule) AddQuestion(question Question) (Question, error) {
	if !surveyModule.ID.Valid() {
		return Question{}, ErrRecordNotFound
	}
	if err := question.Validate(); err != nil {
		return Question{}, err
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return Question{}, dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	question.SetID()
	question.UpdatedAt = time.Now()

	addQuestionChange := mgo.Change{
		Update: bson.M{
			"$push": bson.M{
				"questions": question,
			},
		},
		ReturnNew: false,
	}

	_, err := surveyModuleCol.Find(bson.M{"_id": surveyModule.ID}).Apply(addQuestionChange, nil)
	if err != nil {
		surveyModule.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("UpdateId(%s))", surveyModule.ID),
			"error": err.Error(),
		}).Error(err)
		return Question{}, ErrRecordNotStorable
	}

	if err := surveyModule.UpdateEstimatedDuration(); err != nil {
		return Question{}, err
	}

	return question, nil
}

//UpdateQuestion from SurveyModule
func (surveyModule *SurveyModule) UpdateQuestion(updatedQuestion Question) (Question, error) {
	if !surveyModule.ID.Valid() {
		return Question{}, ErrRecordNotFound
	}
	if !updatedQuestion.ID.Valid() {
		return Question{}, ErrRecordNotFound
	}

	currentQuestion, err := surveyModule.FindQuestionByObjectID(updatedQuestion.ID)
	if err != nil {
		return Question{}, err
	}

	backupedObjectID := currentQuestion.ID
	backupedCreatedAt := currentQuestion.CreatedAt

	currentQuestion.ID = backupedObjectID
	currentQuestion.SPSSID = updatedQuestion.SPSSID
	currentQuestion.CreatedAt = backupedCreatedAt

	currentQuestion.Keyboard = updatedQuestion.Keyboard
	currentQuestion.Explenations = updatedQuestion.Explenations
	currentQuestion.Text = updatedQuestion.Text

	if err := surveyModule.DeleteQuestion(backupedObjectID); err != nil {
		return Question{}, err
	}
	if _, err := surveyModule.AddQuestion(currentQuestion); err != nil {
		return Question{}, err
	}

	return currentQuestion, surveyModule.UpdateQuestionSort()
}

//DeleteQuestion from SurveyModule
func (surveyModule *SurveyModule) DeleteQuestion(questionID bson.ObjectId) error {
	if !surveyModule.ID.Valid() {
		return ErrRecordNotFound
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	deleteQuestionChange := mgo.Change{
		Update: bson.M{
			"$pull": bson.M{
				"questions": bson.M{
					"_id": questionID,
				},
			},
		},
		ReturnNew: false,
	}

	changeInfo, err := surveyModuleCol.Find(bson.M{"_id": surveyModule.ID}).Apply(deleteQuestionChange, nil)
	if err != nil {
		surveyModule.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("UpdateId(%s))", surveyModule.ID),
			"error": err.Error(),
		}).Error(err)
		return ErrRecordNotDeleteable
	}

	if changeInfo.Updated < 1 {
		return ErrRecordNotFound
	}

	return surveyModule.UpdateEstimatedDuration()
}
