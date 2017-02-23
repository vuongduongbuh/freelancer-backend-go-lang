package models

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

//EnsureDocumentExistence checks if a document exists
func EnsureDocumentExistence(collectionName string, documentID bson.ObjectId) error {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	documentCol := dataStore.GetCollection(collectionName)
	defer dataStore.Close()

	documentCount, err := documentCol.FindId(documentID).Count()
	if err != nil {
		getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("Count(%s))", documentID),
		}).Debug(ErrRecordNotRetrievable)
		return ErrRecordNotRetrievable
	}

	if documentCount > 0 {
		return nil
	}

	return ErrRecordNotFound
}
