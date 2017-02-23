package models

import (
	"fmt"
	"time"

	govalidator "gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2/bson"

	"errors"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
)

//CollectionEMailTemplatesProperty mongodb table name
var CollectionEMailTemplatesProperty = "eMailTemplates"

var (
	//ErrEMailTemplateNoTitle template.Title < 1
	ErrEMailTemplateNoTitle = errors.New("email_notitle")
	//ErrEMailTemplateNoBody template.Body < 1
	ErrEMailTemplateNoBody = errors.New("email_nobody")
)

//EMailTemplate model
type EMailTemplate struct {
	BaseModel `bson:",inline"`

	Title       string `json:"title"`
	WebformLink string `json:"webformLink"`

	AuthCodeDescription []LocalizedText `json:"authCodeDescription"`
	Subject             []LocalizedText `json:"subject"`
	Greeting            []LocalizedText `json:"greeting"`
	Button              []LocalizedText `json:"button"`
	SurveyDescription   []LocalizedText `json:"surveyDescription"`
}

//getLogger
func (eMailTemplate *EMailTemplate) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "EMailTemplate",
	})
}

//Validate the given todo item
func (eMailTemplate *EMailTemplate) Validate() error {
	_, err := govalidator.ValidateStruct(eMailTemplate)
	if err != nil {
		return err
	}
	return nil
}

//Save the specific ticket
func (eMailTemplate *EMailTemplate) Save() error {
	if err := eMailTemplate.Validate(); err != nil {
		return err
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	emailTemplatesCol := dataStore.GetCollection(CollectionEMailTemplatesProperty)
	defer dataStore.Close()

	eMailTemplate.SetID()
	eMailTemplate.UpdatedAt = time.Now()

	_, err := emailTemplatesCol.UpsertId(eMailTemplate.ID, bson.M{"$set": eMailTemplate})
	eMailTemplate.getLogger().WithFields(logrus.Fields{
		"query": fmt.Sprintf("UpsertId(%s))", eMailTemplate.ID),
	}).Info(err)
	return err
}

//FindByObjectID the specific template
func (eMailTemplate *EMailTemplate) FindByObjectID(templateID bson.ObjectId) error {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	emailTemplatesCol := dataStore.GetCollection(CollectionEMailTemplatesProperty)
	defer dataStore.Close()

	return emailTemplatesCol.FindId(templateID).One(&eMailTemplate)
}

//GetSurveyDescription returns the body for a language
func (eMailTemplate *EMailTemplate) GetSurveyDescription(language string) string {
	if !eMailTemplate.ID.Valid() {
		return ""
	}

	if len(eMailTemplate.SurveyDescription) < 1 {
		return ""
	}

	for _, body := range eMailTemplate.SurveyDescription {
		if body.Language == language {
			return body.Value
		}
	}

	firstBody := eMailTemplate.SurveyDescription[0]
	return firstBody.Value
}

//GetSubject returns the body for a language
func (eMailTemplate *EMailTemplate) GetSubject(language string) string {
	if !eMailTemplate.ID.Valid() {
		return ""
	}

	if len(eMailTemplate.Subject) < 1 {
		return ""
	}

	for _, body := range eMailTemplate.Subject {
		if body.Language == language {
			return body.Value
		}
	}

	firstBody := eMailTemplate.Subject[0]
	return firstBody.Value
}

//GetGreetingText returns the body for a language
func (eMailTemplate *EMailTemplate) GetGreetingText(language string) string {
	if !eMailTemplate.ID.Valid() {
		return ""
	}

	if len(eMailTemplate.Greeting) < 1 {
		return ""
	}

	for _, body := range eMailTemplate.Greeting {
		if body.Language == language {
			return body.Value
		}
	}

	firstBody := eMailTemplate.Greeting[0]
	return firstBody.Value
}

//GetButtonText returns the button text for language
func (eMailTemplate *EMailTemplate) GetButtonText(language string) string {
	if !eMailTemplate.ID.Valid() {
		return ""
	}

	if len(eMailTemplate.Button) < 1 {
		return ""
	}

	for _, body := range eMailTemplate.Button {
		if body.Language == language {
			return body.Value
		}
	}

	firstBody := eMailTemplate.Button[0]
	return firstBody.Value
}

//GetAuthCodeDescription returns the body for a language
func (eMailTemplate *EMailTemplate) GetAuthCodeDescription(language string) string {
	if !eMailTemplate.ID.Valid() {
		return ""
	}

	if len(eMailTemplate.AuthCodeDescription) < 1 {
		return ""
	}

	for _, body := range eMailTemplate.AuthCodeDescription {
		if body.Language == language {
			return body.Value
		}
	}

	firstBody := eMailTemplate.AuthCodeDescription[0]
	return firstBody.Value
}

//FindAll objects
func (eMailTemplate *EMailTemplate) FindAll() ([]EMailTemplate, error) {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return nil, dataStoreErr
	}
	emailsCol := dataStore.GetCollection(CollectionEMailTemplatesProperty)
	defer dataStore.Close()

	emails := make([]EMailTemplate, 0)
	err := emailsCol.Find(nil).All(&emails)
	if err != nil {
		eMailTemplate.getLogger().WithFields(logrus.Fields{
			"query": "Find(nil)",
			"error": err.Error(),
		}).Debug(ErrRecordsFetch)
		return nil, ErrRecordsFetch
	}

	return emails, nil
}

//FindByID returns a email template
func (eMailTemplate *EMailTemplate) FindByID(id string) error {
	if isObjectID := bson.IsObjectIdHex(id); !isObjectID {
		return ErrInvalidObjectID
	}

	return eMailTemplate.FindByObjectID(bson.ObjectIdHex(id))
}

//Delete email template
func (eMailTemplate *EMailTemplate) Delete() error {
	if !eMailTemplate.ID.Valid() {
		return ErrRecordNotFound
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	emailsCol := dataStore.GetCollection(CollectionEMailTemplatesProperty)
	defer dataStore.Close()

	if err := emailsCol.RemoveId(eMailTemplate.ID); err != nil {
		eMailTemplate.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("RemoveId(%s))", eMailTemplate.ID),
			"error": err.Error(),
		}).Info(err)
		return ErrRecordNotDeleteable
	}

	return nil
}
