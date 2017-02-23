package models

import (
	"os"
	"time"

	"errors"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/mgo.v2"
)

var databaseName string

var (
	//ErrGlobalSessionStoreNoConnection used to specify connection errors
	ErrGlobalSessionStoreNoConnection = errors.New("couldn't establish connection to the database server")
	//ErrGlobalSessionNoGlobalSession if session == nil
	ErrGlobalSessionNoGlobalSession = errors.New("there is no global session saved")
	//ErrGlobalSessionInactive if .Ping err != nil
	ErrGlobalSessionInactive = errors.New("there is no active global session saved")
)

var globalSession *GlobalSessionStore

func getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"controller": "DataStore",
	})
}

//GlobalSessionStore is used to handle the database connection
type GlobalSessionStore struct {
	storedSession *mgo.Session
}

func (globalSessionStore *GlobalSessionStore) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"controller": "GlobalSessionStore",
	})
}

//getDatabaseDialInfo returns the dial infos for mgo
func (globalSessionStore *GlobalSessionStore) getDatabaseDialInfo() *mgo.DialInfo {
	mongoDBHost := os.Getenv("MONGODB_HOST")
	mongoDBName := os.Getenv("MONGODB_NAME")

	if mongoDBHost == "" || mongoDBName == "" {
		globalSessionStore.getLogger().Panic("required db settings missing")
	}

	databaseName = mongoDBName

	dialInfo := &mgo.DialInfo{
		Addrs:    []string{mongoDBHost},
		Database: mongoDBName,
		Timeout:  2 * time.Second,
	}

	mongoDBUsername := os.Getenv("MONGODB_USERNAME")
	if mongoDBUsername != "" {
		dialInfo.Username = mongoDBUsername
	}

	mongoDBPassword := os.Getenv("MONGODB_PASSWORD")
	if mongoDBPassword != "" {
		dialInfo.Password = mongoDBPassword
	}

	return dialInfo
}

//ensureSessionHealth checks the sesion health
func (globalSessionStore *GlobalSessionStore) ensureSessionHealth() error {
	if globalSessionStore.storedSession == nil {
		return ErrGlobalSessionNoGlobalSession
	}
	if err := globalSessionStore.storedSession.Ping(); err != nil {
		return ErrGlobalSessionInactive
	}
	return nil
}

//openGlobalSession opens a new global session
func (globalSessionStore *GlobalSessionStore) openGlobalSession() error {
	mongoSession, err := mgo.DialWithInfo(globalSessionStore.getDatabaseDialInfo())
	if err != nil {
		globalSessionStore.getLogger().WithFields(logrus.Fields{
			"error": err,
		}).Error(ErrGlobalSessionStoreNoConnection.Error())
		return ErrGlobalSessionStoreNoConnection
	}
	globalSessionStore.storedSession = mongoSession
	return nil
}

//GetDataStore returns a new DataStore by .Copy
func (globalSessionStore *GlobalSessionStore) GetDataStore() (*DataStore, error) {
	if err := globalSessionStore.ensureSessionHealth(); err != nil {
		if err := globalSessionStore.openGlobalSession(); err != nil {
			return &DataStore{}, err
		}
	}

	sessionCopy := globalSessionStore.storedSession.Copy()
	if err := sessionCopy.Ping(); err != nil {
		globalSessionStore.getLogger().WithFields(logrus.Fields{
			"function": "GetDataStore",
		}).Fatal(err)
		return &DataStore{}, err
	}

	return &DataStore{sessionCopy}, nil
}

//GetDataStore returns open mongodb session
func GetDataStore() *DataStore {
	if globalSession == nil {
		globalSession = &GlobalSessionStore{}
	}
	dataStore, err := globalSession.GetDataStore()
	if err != nil {
		os.Exit(1)
		panic(ErrGlobalSessionNoGlobalSession)
	}
	return dataStore
}

//CreateDataStore returns open mongodb session or error
func CreateDataStore() (*DataStore, error) {
	if globalSession == nil {
		globalSession = &GlobalSessionStore{}
	}
	return globalSession.GetDataStore()
}

//DataStore is used as general MongoDB session wrapper
type DataStore struct {
	localSession *mgo.Session
}

//GetCollection gets the database adapter
func (dataStore *DataStore) GetCollection(collectionName string) *mgo.Collection {
	return dataStore.localSession.DB(databaseName).C(collectionName)
}

//Close gets the database adapter
func (dataStore *DataStore) Close() {
	dataStore.localSession.Close()
}
