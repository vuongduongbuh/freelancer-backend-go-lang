package models

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	govalidator "gopkg.in/asaskevich/govalidator.v4"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	//ErrParticipantEMailNotUnique is used for duplicated email addresses
	ErrParticipantEMailNotUnique = errors.New("participant_email_duplicate")
)

var (
	//ParticipantDefaultLanguage defines the language if a specific one is not defined
	ParticipantDefaultLanguage = "de"
)

//CollectionParticipantProperty mongodb table name
var CollectionParticipantProperty = "participants"

//Participant model (nested in Company)
type Participant struct {
	BaseModel `bson:",inline"`

	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	EMail     string `json:"email" valid:"email,required"`
	Language  string `json:"language" valid:"required"`

	CompanyID bson.ObjectId `json:"-"`
	IsActive  bool          `json:"isActive" valid:"required"`
}

func (participant Participant) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "Participant",
	})
}

//BootstrapCollection set index
func (participant Participant) BootstrapCollection() error {
	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}

	participantsCol := dataStore.GetCollection(CollectionParticipantProperty)
	defer dataStore.Close()

	index := mgo.Index{
		Key:        []string{"email", "companyid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = participantsCol.EnsureIndex(index)
	if err != nil {
		participant.getLogger().WithFields(logrus.Fields{
			"function": "BootstrapCollection",
			"index":    index,
		}).Error(err)
		return err
	}

	return nil
}

//Validate the current participant
func (participant *Participant) Validate() error {
	_, err := govalidator.ValidateStruct(participant)
	if err != nil {
		return err
	}
	return nil
}

//Create sets the defaults and generates an ID
func (participant *Participant) Create() error {
	participant.SetID()
	participant.UpdatedAt = time.Now()

	participant.Language = ParticipantDefaultLanguage
	participant.IsActive = true

	return nil
}

//Save saves the participant
func (participant *Participant) Save() error {
	if err := participant.Validate(); err != nil {
		return err
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	participantsCol := dataStore.GetCollection(CollectionParticipantProperty)
	defer dataStore.Close()

	_, err := participantsCol.UpsertId(participant.ID, bson.M{"$set": participant})
	return err
}

//FindParticipantByObjectID finds participant by ObjectID
func (participant *Participant) FindParticipantByObjectID(participantID bson.ObjectId) error {
	if !participantID.Valid() {
		return ErrInvalidObjectID
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	participantsCol := dataStore.GetCollection(CollectionParticipantProperty)
	defer dataStore.Close()

	err := participantsCol.FindId(participantID).One(&participant)
	return err
}

//FindParticipantInCompanyByObjectID finds participant in a given company by ObjectID
func (participant *Participant) FindParticipantInCompanyByObjectID(companyID bson.ObjectId, participantID bson.ObjectId) error {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	participantsCol := dataStore.GetCollection(CollectionParticipantProperty)
	defer dataStore.Close()

	err := participantsCol.Find(bson.M{
		"_id":       participantID,
		"companyid": companyID,
	}).One(&participant)
	return err
}

//FindParticipantsByCompany finds participant from a company
func (participant *Participant) FindParticipantsByCompany(companyID bson.ObjectId) ([]Participant, error) {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return nil, dataStoreErr
	}
	participantsCol := dataStore.GetCollection(CollectionParticipantProperty)
	defer dataStore.Close()

	participants := make([]Participant, 0)
	err := participantsCol.Find(bson.M{
		"companyid": companyID,
	}).All(&participants)

	if err != nil {
		participant.getLogger().WithFields(logrus.Fields{
			"function":  "FindParticipantsByCompany",
			"companyID": companyID,
		}).Info(err)
		return participants, ErrRecordNotFound
	}

	return participants, nil
}

//BulkAddParticipantsToCompany bulk adds participants by company
func (participant *Participant) BulkAddParticipantsToCompany(companyID bson.ObjectId, participants []Participant) error {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	participantsCol := dataStore.GetCollection(CollectionParticipantProperty)
	defer func() {
		if r := recover(); r != nil {
			participant.getLogger().WithFields(logrus.Fields{
				"function": "BulkAddParticipantsToCompany",
			}).Info(errors.New("participants_toomanyduplicates"))
		}
		dataStore.Close()
	}()

	mgoBulkConn := participantsCol.Bulk()
	mgoBulkConn.Unordered()

	for _, participant := range participants {
		participant.CompanyID = companyID
		if err := participant.Validate(); err != nil {
			participant.getLogger().WithFields(logrus.Fields{
				"function":    "BulkAddParticipantsToCompany",
				"participant": participant,
			}).Info(err)
			continue
		}
		mgoBulkConn.Insert(participant)
	}

	if _, err := mgoBulkConn.Run(); err != nil {
		participant.getLogger().WithFields(logrus.Fields{
			"function": "BulkAddParticipantsToCompany",
			"action":   "mgoBulkConn.Run()",
		}).Info(err)
		return ErrParticipantEMailNotUnique
	}
	return nil
}

//Delete a participant if found
func (participant *Participant) Delete(companyID bson.ObjectId) error {
	if !companyID.Valid() {
		return ErrRecordNotFound
	}
	if !participant.ID.Valid() {
		return ErrRecordNotFound
	}
	return participant.DeleteByObjectID(companyID, participant.ID)
}

//DeleteByObjectID a participant without fetching it into the RAM
func (participant Participant) DeleteByObjectID(companyID bson.ObjectId, participantID bson.ObjectId) error {
	if !companyID.Valid() {
		return ErrRecordNotFound
	}
	if !participantID.Valid() {
		return ErrRecordNotFound
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	participantCol := dataStore.GetCollection(CollectionParticipantProperty)
	defer dataStore.Close()

	participantQuery := bson.M{
		"_id":       participantID,
		"companyid": companyID,
	}

	if err := participantCol.Remove(participantQuery); err != nil {
		participant.getLogger().WithFields(logrus.Fields{
			"function":    "Delete",
			"query":       participantQuery,
			"participant": participant,
		}).Info(err)
		return ErrRecordNotStorable
	}

	return nil
}
