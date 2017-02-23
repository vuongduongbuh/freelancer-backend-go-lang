package models

import "errors"

var (
	ErrRecordsFetch                  = errors.New("couldn't fetch items from collection")
	ErrRecordNotFound                = errors.New("can't find item with given id")
	ErrRecordNotRetrievable          = errors.New("can't fetch item")
	ErrRecordNotDeleteable           = errors.New("can't delete item")
	ErrRecordNotStorable             = errors.New("can't save item to db")
	ErrDuplicatedIndex               = errors.New("can't save item to db dut to a duplicated index")
	ErrInvalidObjectID               = errors.New("id is not a valid hex id")
	ErrGeneralDatabaseException      = errors.New("there was a problem working with the db")
	ErrGeneralDatabaseIndexException = errors.New("couldn't generate an index with the given params")
	ErrEmptyRequestBody              = errors.New("request body is empty")
	ErrInvalidToken                  = errors.New("auth_invalidtoken")
	ErrQuestionsCountMissmatch       = errors.New("questions_countmissmatch")
)
