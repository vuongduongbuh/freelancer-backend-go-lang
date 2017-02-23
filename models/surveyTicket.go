package models

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	govalidator "gopkg.in/asaskevich/govalidator.v4"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
)

//CollectionTicketsProperty mongodb table name
var CollectionTicketsProperty = "tickets"

var (
	//SurveyTicketsAuthCodeLength authcode length
	SurveyTicketsAuthCodeLength = 6
)

var (
	//SurveyTicketPushTokenTypeIOS used for ios tokens
	SurveyTicketPushTokenTypeIOS = "ios"
	//SurveyTicketPushTokenTypeGCM used for android tokens
	SurveyTicketPushTokenTypeGCM = "android"
)

var (
	//ErrDuplicatedAuthCode is returned on insert index errors
	ErrDuplicatedAuthCode = errors.New("surveytickets_dupauthcode")
	//ErrDuplicatedTickets too many duplicates
	ErrDuplicatedTickets = errors.New("surveytickets_toomanyduplicates")
	//ErrSurveyIncomplete is sent back if /submit is called before finished
	ErrSurveyIncomplete = errors.New("survey_surveyincomplete")
)

//SurveyTicketConductRequestModel model for /conduct
type SurveyTicketConductRequestModel struct {
	QuestionID       bson.ObjectId `json:"questionId" valid:"required"`
	Answer           string        `json:"answer"`
	MeasuredDuration float32       `json:"measuredDuration"`
}

//SurveyTicketConductRequestModelQuickFix model for /conduct
type SurveyTicketConductRequestModelQuickFix struct {
	QuestionID       bson.ObjectId `json:"questionId" valid:"required"`
	Answer           string        `json:"answer"`
	MeasuredDuration interface{}   `json:"measuredDuration"`
}

//SurveyTicketSentNotificationModel model
type SurveyTicketSentNotificationModel struct {
	NotificationObjectID bson.ObjectId `json:"notificationObjectID"`
	TodoObjectID         bson.ObjectId `json:"todoObjectID"`
	SentAt               time.Time     `json:"sentAt"`
	DeliveryType         int           `json:"deliveryType"`
	PushToken            string        `json:"pushToken,omitempty"  bson:",omitempty"`
	EMail                string        `json:"eMail,omitempty"  bson:",omitempty"`
}

//SurveyTicket model
type SurveyTicket struct {
	BaseModel `bson:",inline"`

	SurveyCatalogID bson.ObjectId `json:"surveyCatalogID" valid:"required"`
	Company         Company       `json:"company" valid:"required"`
	Participant     Participant   `json:"participant" valid:"required"`

	AuthCode string `json:"authCode" valid:"required"`

	IsAuthCodeSent bool      `json:"isAuthCodeSent"`
	AuthCodeSentAt time.Time `json:"authCodeSentAt"`

	Todos []SurveyTicketTodo `json:"todos" valid:"required"`

	IsActive  bool      `json:"isActive"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`

	PushToken     string `json:"pushToken,omitempty" bson:",omitempty"`
	PushTokenType string `json:"pushTokenType,omitempty" bson:",omitempty"`

	SentNotifications []SurveyTicketSentNotificationModel `json:"sentNotifications"`
}

//getLogger
func (surveyTicket *SurveyTicket) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "SurveyTicket",
	})
}

//BootstrapCollection set index
func (surveyTicket SurveyTicket) BootstrapCollection() error {
	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}

	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	authCodeIndex := mgo.Index{
		Key:        []string{"authcode"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = ticketsCol.EnsureIndex(authCodeIndex)
	if err != nil {
		surveyTicket.getLogger().WithFields(logrus.Fields{
			"function": "BootstrapCollection",
			"index":    authCodeIndex,
		}).Error(err)
		return err
	}

	return nil
}

//Validate the given todo item
func (surveyTicket *SurveyTicket) Validate() error {
	_, err := govalidator.ValidateStruct(surveyTicket)
	if err != nil {
		return err
	}
	return nil
}

//Save the specific ticket
func (surveyTicket *SurveyTicket) Save() error {
	if err := surveyTicket.Validate(); err != nil {
		return err
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	surveyTicket.UpdatedAt = time.Now()

	if _, err := ticketsCol.UpsertId(surveyTicket.ID, bson.M{"$set": surveyTicket}); err != nil {
		surveyTicket.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("UpsertId(%s))", surveyTicket.ID),
		}).Error(err)
		return err
	}

	return nil
}

//SetUserSettings the specific push token to ticket
func (surveyTicket *SurveyTicket) SetUserSettings(pushToken string, pushTokenType string, language string) error {
	if err := surveyTicket.Validate(); err != nil {
		return err
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	surveyTicket.UpdatedAt = time.Now()

	mgoUserSettingsChangeSet := bson.M{
		"updatedAt": surveyTicket.UpdatedAt,
	}

	if pushToken != "" {
		mgoUserSettingsChangeSet["pushtoken"] = pushToken
	}
	if pushTokenType != "" {
		mgoUserSettingsChangeSet["pushtokentype"] = pushTokenType
	}
	if language != "" {
		mgoUserSettingsChangeSet["participant.language"] = language
	}

	addUserSettingsChange := mgo.Change{
		Update: bson.M{
			"$set": mgoUserSettingsChangeSet,
		},
		ReturnNew: false,
	}

	_, err := ticketsCol.Find(bson.M{"_id": surveyTicket.ID}).Apply(addUserSettingsChange, nil)
	if err != nil {
		surveyTicket.getLogger().Info(err)
		return ErrRecordNotDeleteable
	}
	return err
}

//DeletePushToken the specific push token in ticket
func (surveyTicket *SurveyTicket) DeletePushToken() error {
	if err := surveyTicket.Validate(); err != nil {
		return err
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	surveyTicket.UpdatedAt = time.Now()

	addPushTokenChange := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"pushtoken": "",
				"updatedAt": surveyTicket.UpdatedAt,
			},
		},
		ReturnNew: false,
	}

	_, err := ticketsCol.Find(bson.M{"_id": surveyTicket.ID}).Apply(addPushTokenChange, nil)
	if err != nil {
		surveyTicket.getLogger().Info(err)
		return ErrRecordNotDeleteable
	}
	return err
}

//FindByID returns a surveyTicketID
func (surveyTicket *SurveyTicket) FindByID(id string) error {
	if isObjectID := bson.IsObjectIdHex(id); !isObjectID {
		return ErrInvalidObjectID
	}

	return surveyTicket.FindByObjectID(bson.ObjectIdHex(id))
}

//FindByObjectID returns a surveyTicketID
func (surveyTicket *SurveyTicket) FindByObjectID(surveyTicketID bson.ObjectId) error {
	if isValid := surveyTicketID.Valid(); !isValid {
		return ErrInvalidObjectID
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	if err := ticketsCol.FindId(surveyTicketID).One(&surveyTicket); err != nil {
		surveyTicket.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("FindId(%s))", surveyTicketID),
		}).Info(err)
		return ErrRecordNotRetrievable
	}

	return nil
}

//FindByAuthHeader is used to return a ticket by auth header
func (surveyTicket *SurveyTicket) FindByAuthHeader(rawAuthHeader string) error {
	decodedAuthHeader, err := base64.URLEncoding.DecodeString(rawAuthHeader)
	if err != nil {
		return ErrInvalidToken
	}
	authHeaderFragments := strings.Split(string(decodedAuthHeader), "::")

	if len(authHeaderFragments) < 1 {
		return ErrInvalidToken
	}

	authCode := authHeaderFragments[0]

	surveyTicketQuery := bson.M{
		"authcode": authCode,
		"isactive": true,
	}

	if len(authHeaderFragments) == 2 {
		participantEMail := authHeaderFragments[1]
		if isEmail := govalidator.IsEmail(participantEMail); !isEmail {
			return ErrInvalidToken
		}
		surveyTicketQuery["participant.email"] = participantEMail
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}
	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	return ticketsCol.Find(surveyTicketQuery).One(&surveyTicket)
}

//GenerateAuthCode generates the specific auth code
func (surveyTicket *SurveyTicket) GenerateAuthCode() {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, SurveyTicketsAuthCodeLength)
	for i := 0; i < SurveyTicketsAuthCodeLength; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	surveyTicket.AuthCode = string(result)
}

//AnswerQuestions is used to answer questions nested into a todo
func (surveyTicket *SurveyTicket) AnswerQuestions(todoID string, questionAnswers []SurveyTicketConductRequestModel, applicationIdentifier string) error {
	if !surveyTicket.ID.Valid() {
		return ErrInvalidObjectID
	}
	if isObjectID := bson.IsObjectIdHex(todoID); !isObjectID {
		return ErrInvalidObjectID
	}

	questionAnswersMap := make(map[bson.ObjectId]SurveyTicketConductRequestModel, 0)

	//validate each sent answer
	for _, questionAnswer := range questionAnswers {
		_, err := govalidator.ValidateStruct(questionAnswer)
		if err != nil {
			return err
		}
		questionAnswersMap[questionAnswer.QuestionID] = questionAnswer
	}

	var foundTodo SurveyTicketTodo
	for todoIndex, todo := range surveyTicket.Todos {
		if todo.ID != bson.ObjectIdHex(todoID) || todo.Status != SurveyTodoStateActive {
			continue
		}
		foundTodo = todo

		questionCount := float32(len(todo.Questions))
		questionAnsweredCount := 0
		questionsMeasuredDuration := float32(0)

		for questionIndex, question := range todo.Questions {
			questionAnswer, isAnsweredQuestion := questionAnswersMap[question.ID]
			if !isAnsweredQuestion {
				if question.MeasuredDuration > 0 {
					questionAnsweredCount++
					questionsMeasuredDuration += questionAnswer.MeasuredDuration
				}
				continue
			}

			question.Answer = questionAnswer.Answer
			question.MeasuredDuration = questionAnswer.MeasuredDuration
			question.ApplicationIdentifier = applicationIdentifier
			question.AnsweredAt = time.Now()

			questionsMeasuredDuration += questionAnswer.MeasuredDuration

			surveyTicket.Todos[todoIndex].Questions[questionIndex] = question
			surveyTicket.Todos[todoIndex].UpdatedAt = time.Now()

			questionAnsweredCount++
		}

		if questionCount > 0 && questionAnsweredCount > 0 {
			surveyTicket.Todos[todoIndex].Progress = (100 / questionCount) * float32(questionAnsweredCount)
			surveyTicket.Todos[todoIndex].MeasuredDuration = questionsMeasuredDuration
		}
	}

	if !foundTodo.ID.Valid() {
		return ErrRecordNotFound
	}

	return surveyTicket.Save()
}

//GetAnsweredQuestions is used to return already answered questions
func (surveyTicket *SurveyTicket) GetAnsweredQuestions(todoID string) ([]SurveyTicketConductRequestModel, error) {
	if !surveyTicket.ID.Valid() {
		return nil, ErrInvalidObjectID
	}
	if isObjectID := bson.IsObjectIdHex(todoID); !isObjectID {
		return nil, ErrInvalidObjectID
	}
	todoObjectID := bson.ObjectIdHex(todoID)

	foundTodo := false
	questionAnswers := make([]SurveyTicketConductRequestModel, 0)
	for _, todo := range surveyTicket.Todos {
		if todo.ID != todoObjectID {
			continue
		}
		foundTodo = true

		for _, question := range todo.Questions {
			if question.MeasuredDuration != 0 {
				questionAnswer := SurveyTicketConductRequestModel{
					QuestionID:       question.ID,
					Answer:           question.Answer,
					MeasuredDuration: question.MeasuredDuration,
				}
				questionAnswers = append(questionAnswers, questionAnswer)
			}
		}
	}

	if !foundTodo {
		return nil, ErrRecordNotFound
	}

	return questionAnswers, nil
}

//SubmitTodo marks todo as done
func (surveyTicket *SurveyTicket) SubmitTodo(todoID string, applicationIdentifier string) error {
	if !surveyTicket.ID.Valid() {
		return ErrInvalidObjectID
	}
	if isObjectID := bson.IsObjectIdHex(todoID); !isObjectID {
		return ErrInvalidObjectID
	}

	todoObjectID := bson.ObjectIdHex(todoID)
	foundTodo := false
	for todoID, todo := range surveyTicket.Todos {
		if todo.ID != todoObjectID {
			continue
		}
		foundTodo = true

		for _, question := range todo.Questions {
			if question.MeasuredDuration == 0 && question.Keyboard.Type != QuestionKeyboardTypeInformation {
				return ErrSurveyIncomplete
			}
		}

		if applicationIdentifier != "" {
			surveyTicket.Todos[todoID].SubmittedAt = time.Now()
			surveyTicket.Todos[todoID].ApplicationIdentifier = applicationIdentifier
		}

		surveyTicket.Todos[todoID].Status = SurveyTodoStateDone
	}

	if !foundTodo {
		return ErrRecordNotFound
	}

	return surveyTicket.Save()
}

//Activate marks ticket as active
func (surveyTicket SurveyTicket) Activate() error {
	if !surveyTicket.ID.Valid() {
		return ErrInvalidObjectID
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}

	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	setIsActiveTrue := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"isactive":  true,
				"updatedAt": surveyTicket.UpdatedAt,
			},
		},
		ReturnNew: false,
	}

	surveyTicketIDQuery := bson.M{
		"_id": surveyTicket.ID,
	}

	_, err = ticketsCol.Find(surveyTicketIDQuery).Apply(setIsActiveTrue, nil)
	if err != nil {
		surveyTicket.getLogger().Info(err)
		return ErrRecordNotDeleteable
	}

	return nil
}

//SetAuthCodeSent marks ticket as auth code sent
func (surveyTicket SurveyTicket) SetAuthCodeSent() error {
	if !surveyTicket.ID.Valid() {
		return ErrInvalidObjectID
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}

	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	now := time.Now()

	setAuchCodeSentTrue := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"isauthcodesent": true,
				"authcodesentat": now,
				"isactive":       true,
				"updatedAt":      now,
			},
		},
		ReturnNew: false,
	}

	surveyTicketIDQuery := bson.M{
		"_id": surveyTicket.ID,
	}

	_, err = ticketsCol.Find(surveyTicketIDQuery).Apply(setAuchCodeSentTrue, nil)
	if err != nil {
		surveyTicket.getLogger().Info(err)
		return ErrRecordNotDeleteable
	}

	return nil
}

//EnrichActiveTodos adds the questions to all active todos
func (surveyTicket SurveyTicket) EnrichActiveTodos() error {
	if !surveyTicket.ID.Valid() {
		return ErrInvalidObjectID
	}

	hasActiveTodo := false
	for _, todo := range surveyTicket.Todos {
		if todo.Status != SurveyTodoStateActive {
			continue
		}
		if !todo.IsEnriched {
			hasActiveTodo = true
		}
	}

	if !hasActiveTodo {
		return nil
	}

	var surveyCatalog SurveyCatalog
	if err := surveyCatalog.FindByObjectID(surveyTicket.SurveyCatalogID); err != nil {
		return err
	}

	mapedRichTodos := make(map[bson.ObjectId]SurveyTicketTodo)
	for _, todo := range surveyCatalog.TicketTodos {
		mapedRichTodos[todo.ID] = todo
	}

	for todoIdx, todo := range surveyTicket.Todos {
		if todo.Status != SurveyTodoStateActive {
			continue
		}

		richTodo, doesExsist := mapedRichTodos[todo.ID]
		if !doesExsist {
			surveyTicket.getLogger().WithFields(logrus.Fields{
				"function": "EnrichActiveTodos",
				"todoID":   todo.ID,
			}).Fatal("Todo in ticket can't be found inside surveyCatalog.TicketTodos")
			return ErrRecordNotFound
		}

		richTodo.Status = todo.Status
		richTodo.IsEnriched = true
		richTodo.EnrichedAt = time.Now()
		surveyTicket.Todos[todoIdx] = richTodo
	}

	return surveyTicket.Save()
}

//IsNotificationSent checks if a notification is sent
func (surveyTicket *SurveyTicket) IsNotificationSent(notificationID, surveyTodoID bson.ObjectId) bool {
	for _, sentNotification := range surveyTicket.SentNotifications {
		if sentNotification.NotificationObjectID == notificationID && sentNotification.TodoObjectID == surveyTodoID {
			return true
		}
	}
	return false
}