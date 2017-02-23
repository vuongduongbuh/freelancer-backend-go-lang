package models

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2/bson"
)

const (
	QuestionKeyboardTypeDefault     = "default"
	QuestionKeyboardTypeSmileboard  = "smileboard"
	QuestionKeyboardTypeText        = "text"
	QuestionKeyboardTypeInformation = "information"
)

//LocalizedText model
type LocalizedText struct {
	Language string `json:"language" valid:"required"`
	Value    string `json:"value" valid:"required"`
}

//QuestionKeyboardKey model
type QuestionKeyboardKey struct {
	Value string          `json:"value" valid:"required"`
	Text  []LocalizedText `json:"text" bson:",omitempty" valid:"required"`
}

//QuestionKeyboard model
type QuestionKeyboard struct {
	Type                 string                `json:"type" valid:"keyboardType,required"`
	Keys                 []QuestionKeyboardKey `json:"keys,omitempty" bson:",omitempty" valid:"keyboardKeys,optional"`
	Title                []LocalizedText       `json:"title,omitempty" bson:",omitempty"`
	Button               []LocalizedText       `json:"button,omitempty" bson:",omitempty"`
	Placeholder          []LocalizedText       `json:"placeholder,omitempty" bson:",omitempty"`
	Image                string                `json:"image,omitempty" bson:",omitempty"`
	EnableMultipleChoice bool                  `json:"enableMultipleChoice,omitempty" bson:",omitempty"`
	Min                  int                   `json:"min"`
	Max                  int                   `json:"max,omitempty" bson:",omitempty"`
	Unit                 string                `json:"unit,omitempty" bson:",omitempty"`
	IsOptional           bool                  `json:"isOptional"`
}

//QuestionExplenation model
type QuestionExplenation struct {
	Language string `json:"language" valid:"required"`
	Title    string `json:"title" valid:"required"`
	Text     string `json:"text" valid:"required"`
}

//Question model
type Question struct {
	BaseModel `bson:",inline"`

	SPSSID string `json:"spssId" valid:"spssid,required"`

	Text     []LocalizedText  `json:"text" bson:",omitempty"`
	Keyboard QuestionKeyboard `json:"keyboard" valid:"questionKeyboard,required"`

	Explenations []QuestionExplenation `json:"explenations" valid:"questionExplenations,optional"`

	EstimatedDuration float32 `json:"estimatedDuration"`

	MeasuredDuration      float32   `json:"measuredDuration,omitempty" bson:",omitempty"`
	Answer                string    `json:"answer,omitempty" bson:",omitempty"`
	AnsweredAt            time.Time `json:"answeredAt,omitempty" bson:",omitempty"`
	ApplicationIdentifier string    `json:"applicationIdentifier,omitempty" bson:",omitempty"`

	SortIndex int `json:"sortIndex"`
}

//Validate the given question
func (question *Question) Validate() error {
	_, err := govalidator.ValidateStruct(question)
	if err != nil {
		return err
	}
	return nil
}

//IsSPSSIDUnique checks the given SPSSID
func (question *Question) IsSPSSIDUnique() bool {
	if !question.ID.Valid() {
		question.SetID()
	}

	if question.SPSSID == "" {
		return false
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return false
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyModuleProperty)
	defer dataStore.Close()

	spssIDQuery := bson.M{
		"questions.spssid": question.SPSSID,
		"questions._id": bson.M{
			"$ne": question.ID,
		},
	}

	count, err := surveyModuleCol.Find(spssIDQuery).Count()
	if err != nil {
		helpers.GetLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("Find(%s))", spssIDQuery),
		}).Info(err)
		return false
	}

	return count < 1
}
