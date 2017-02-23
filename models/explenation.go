package models

import (
	"fmt"
	"time"

	"gitlab.com/slsurvey/slsurvey-srv/helpers"

	"github.com/Sirupsen/logrus"

	govalidator "gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2/bson"
)

//CollectionExplenationsProperty mogodb table name
var CollectionExplenationsProperty = "explenations"

//ExplenationContent model
type ExplenationContent struct {
	Language string `json:"language" valid:"required"`
	Title    string `json:"title" valid:"required"`
	Text     string `json:"text" valid:"required"`
}

//Explenation model
type Explenation struct {
	BaseModel `bson:",inline"`

	Content []ExplenationContent `json:"content" valid:"explenationContent,required"`
}

func (explenation Explenation) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model":      "Explenation",
		"collection": CollectionExplenationsProperty,
	})
}

//FindAll explenations
func (explenation *Explenation) FindAll() ([]Explenation, error) {
	dataStore, err := CreateDataStore()
	if err != nil {
		return nil, err
	}
	explenationsCol := dataStore.GetCollection(CollectionExplenationsProperty)
	defer dataStore.Close()

	explenations := make([]Explenation, 0)
	err = explenationsCol.Find(nil).All(&explenations)
	if err != nil {
		explenation.getLogger().WithFields(logrus.Fields{
			"query": "Find(nil)",
			"error": err.Error(),
		}).Debug(err)
		return nil, ErrRecordsFetch
	}

	return explenations, nil
}

//FindByID returns a explenation
func (explenation *Explenation) FindByID(id string) error {
	if isObjectID := bson.IsObjectIdHex(id); !isObjectID {
		return helpers.ErrInvalidObjectID
	}
	return explenation.FindByObjectID(bson.ObjectIdHex(id))
}

//FindByObjectID returns a explenation
func (explenation *Explenation) FindByObjectID(explenationID bson.ObjectId) error {
	if isValid := explenationID.Valid(); !isValid {
		return helpers.ErrInvalidObjectID
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}
	explenationsCol := dataStore.GetCollection(CollectionExplenationsProperty)
	defer dataStore.Close()

	if err := explenationsCol.Find(bson.M{"_id": explenationID}).One(&explenation); err != nil {
		explenation.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("FindId(%s))", explenationID),
			"error": err.Error(),
		}).Info(err)
		return helpers.ErrRecordNotFound
	}

	return nil
}

//Save or updates the given object
func (explenation *Explenation) Save() error {
	_, err := govalidator.ValidateStruct(explenation)
	if err != nil {
		return err
	}

	dataStore, dataStoreErr := CreateDataStore()
	if err != nil {
		return dataStoreErr
	}
	explenationsCol := dataStore.GetCollection(CollectionExplenationsProperty)
	defer dataStore.Close()

	explenation.SetID()
	explenation.UpdatedAt = time.Now()

	_, err = explenationsCol.UpsertId(explenation.ID, bson.M{"$set": explenation})
	return err
}

//Delete explenation
func (explenation *Explenation) Delete() error {
	if !explenation.ID.Valid() {
		return ErrRecordNotFound
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	explenationsCol := dataStore.GetCollection(CollectionExplenationsProperty)
	defer dataStore.Close()

	if err := explenationsCol.RemoveId(explenation.ID); err != nil {
		explenation.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("RemoveId(%s))", explenation.ID),
			"error": err.Error(),
		}).Info(err)
		return ErrRecordNotDeleteable
	}

	return nil
}
