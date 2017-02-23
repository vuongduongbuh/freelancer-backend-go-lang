package models

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2/bson"
)

//CollectionCompanyProperty mongodb table name
var CollectionCompanyProperty = "companies"

//CompanyStyles to style the applications
type CompanyStyles struct {
	PrimaryColorCode string `json:"primaryColorCode" valid:"required"`
	Logo             string `json:"logo" bson:",omitempty"`
}

//Company model
type Company struct {
	BaseModel `bson:",inline"`

	Name      string  `json:"name" valid:"required"`
	Headcount float32 `json:"headcount"`
	Industry  string  `json:"industry" valid:"required"`

	IsActive      bool          `json:"isActive"`
	CompanyStyles CompanyStyles `json:"companyStyles" bson:",omitempty"`
}

func (company Company) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model":      "Company",
		"collection": CollectionCompanyProperty,
	})
}

//Save or updates the given object
func (company *Company) Save() error {
	_, err := govalidator.ValidateStruct(company)
	if err != nil {
		return err
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}
	companiesCol := dataStore.GetCollection(CollectionCompanyProperty)
	defer dataStore.Close()

	company.SetID()
	company.UpdatedAt = time.Now()

	_, err = companiesCol.UpsertId(company.ID, bson.M{"$set": company})
	return err
}

//FindAll objects
func (company *Company) FindAll() ([]Company, error) {
	dataStore, err := CreateDataStore()
	if err != nil {
		return nil, err
	}
	companiesCol := dataStore.GetCollection(CollectionCompanyProperty)
	defer dataStore.Close()

	companies := make([]Company, 0)
	err = companiesCol.Find(nil).All(&companies)
	if err != nil {
		company.getLogger().WithFields(logrus.Fields{
			"query": "Find(nil)",
			"error": err.Error(),
		}).Debug(ErrRecordsFetch)
		return nil, ErrRecordsFetch
	}

	return companies, nil
}

//FindByID returns a company
func (company *Company) FindByID(id string) error {
	if isObjectID := bson.IsObjectIdHex(id); !isObjectID {
		return ErrInvalidObjectID
	}

	return company.FindByObjectID(bson.ObjectIdHex(id))
}

//FindByObjectID returns a company
func (company *Company) FindByObjectID(companyID bson.ObjectId) error {
	if isValid := companyID.Valid(); !isValid {
		return ErrInvalidObjectID
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}
	companiesCol := dataStore.GetCollection(CollectionCompanyProperty)
	defer dataStore.Close()

	if err := companiesCol.Find(bson.M{"_id": companyID}).One(&company); err != nil {
		company.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("FindId(%s))", companyID),
			"error": err.Error(),
		}).Info(err)
		return ErrRecordNotRetrievable
	}

	return nil
}

//Delete company
func (company *Company) Delete() error {
	if !company.ID.Valid() {
		return ErrRecordNotFound
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	companiesCol := dataStore.GetCollection(CollectionCompanyProperty)
	defer dataStore.Close()

	if err := companiesCol.RemoveId(company.ID); err != nil {
		company.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("RemoveId(%s))", company.ID),
			"error": err.Error(),
		}).Info(err)
		return ErrRecordNotDeleteable
	}

	return nil
}

//FindParticipants returns all participants of the given company
func (company *Company) FindParticipants() ([]Participant, error) {
	if !company.ID.Valid() {
		return []Participant{}, ErrRecordNotFound
	}

	var participant Participant
	return participant.FindParticipantsByCompany(company.ID)
}

//FindParticipantByObjectID returns a single participant
func (company *Company) FindParticipantByObjectID(participantID bson.ObjectId) (Participant, error) {
	if !company.ID.Valid() {
		return Participant{}, ErrRecordNotFound
	}
	if !participantID.Valid() {
		return Participant{}, ErrRecordNotFound
	}

	var participant Participant
	if err := participant.FindParticipantInCompanyByObjectID(company.ID, participantID); err != nil {
		return Participant{}, err
	}

	return participant, nil
}

//AddParticipant ensure the email is unique and adds the record to the participant array
func (company *Company) AddParticipant(participant Participant) (Participant, error) {
	participant.SetID()
	participant.UpdatedAt = time.Now()
	participant.CompanyID = company.ID

	if err := participant.Save(); err != nil {
		company.getLogger().Error(err)
		return Participant{}, ErrRecordNotStorable
	}

	return participant, nil
}

//BulkAddParticipants ensure the email is unique and adds the record to the participant array
func (company *Company) BulkAddParticipants(participants []Participant) error {
	if !company.ID.Valid() {
		return ErrRecordNotFound
	}

	var participant Participant
	return participant.BulkAddParticipantsToCompany(company.ID, participants)
}

//DeleteParticipant removes the participant from the given company
func (company *Company) DeleteParticipant(participantID bson.ObjectId) error {
	if !company.ID.Valid() {
		return ErrRecordNotFound
	}
	if !participantID.Valid() {
		return ErrRecordNotFound
	}

	var participant Participant
	return participant.DeleteByObjectID(company.ID, participantID)
}

//UpdateParticipant updates a participant from the given company
func (company *Company) UpdateParticipant(changedParticipant Participant) error {
	if !company.ID.Valid() {
		return ErrRecordNotFound
	}
	if !changedParticipant.ID.Valid() {
		return ErrRecordNotFound
	}
	return changedParticipant.Save()
}
