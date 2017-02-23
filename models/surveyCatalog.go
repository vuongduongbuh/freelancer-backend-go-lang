package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2/bson"
)

//CollectionSurveyCatalogProperty mogodb table name
var CollectionSurveyCatalogProperty = "surveyCatalogs"

var (
	//ErrSurveyCatalogAlreadyPublished survey catalog is already published (isDraft = false) and thus can not be edited
	ErrSurveyCatalogAlreadyPublished = errors.New("sc_alreadypublished")
	//ErrSurveyCatalogStillActive survey catalog is already published (isDone != true) and thus can not be edited
	ErrSurveyCatalogStillActive = errors.New("sc_alreadypublished")
	//ErrSurveyCatalogStartDateAfterEndDate startDate > endDate
	ErrSurveyCatalogStartDateAfterEndDate = errors.New("sc_startdateafterenddate")
	//ErrSurveyCatalogStartDateBeforeEndDate startDate > endDate
	ErrSurveyCatalogStartDateBeforeEndDate = errors.New("sc_startdatebeforeenddate")
	//ErrSurveyCatalogStartDateInPast startDate < now
	ErrSurveyCatalogStartDateInPast = errors.New("sc_startdateinpast")
	//ErrSurveyCatalogNotificationBeforeStart notificationDate < startDate
	ErrSurveyCatalogNotificationBeforeStart = errors.New("sc_notificationbeforestart")
	//ErrSurveyCatalogNotificationAfterEnd notificationDate > startDate
	ErrSurveyCatalogNotificationAfterEnd = errors.New("sc_notificationafterend")
	//ErrSurveyCatalogNotificationBeforeLast notificationDate < notificationDate-1
	ErrSurveyCatalogNotificationBeforeLast = errors.New("sc_notificationbeforelast")
)

//SurveyCatalogAddSettingsGeneralSettingsModel is used to set up the base and end survey
type SurveyCatalogAddSettingsGeneralSettingsModel struct {
	Notifications []time.Time `json:"notifications" valid:"required"`
	StartDate     time.Time   `json:"startDate" valid:"required"`
	EndDate       time.Time   `json:"endDate" valid:"required"`
}

//SurveyCatalogAddSettingsRecurringSettingsModel is used to set up the recurring (weekly) survey
type SurveyCatalogAddSettingsRecurringSettingsModel struct {
	SurveyCatalogAddSettingsGeneralSettingsModel

	Repeats int `json:"repeats" valid:"required"`
}

//SurveyCatalogAddSettingsRequestModel is used to set up the catalog
type SurveyCatalogAddSettingsRequestModel struct {
	Base      SurveyCatalogAddSettingsGeneralSettingsModel   `json:"base" valid:"required"`
	Recurring SurveyCatalogAddSettingsRecurringSettingsModel `json:"recurring,omitempty" bson:",omitempty"`
	End       SurveyCatalogAddSettingsGeneralSettingsModel   `json:"end" valid:"required"`
}

var (
	//SurveyCatalogTransportTypeEmail force email delivery
	SurveyCatalogTransportTypeEmail = 1
	//SurveyCatalogTransportTypePush force push delivery
	SurveyCatalogTransportTypePush = 2
)

//SurveyCatalogNotification is used to store notifications
type SurveyCatalogNotification struct {
	ID bson.ObjectId `json:"id" bson:"_id,omitempty"`

	Date          time.Time `json:"date"`
	TransportType int       `json:"transportType"`

	IsSent bool      `json:"isSent"`
	SentAt time.Time `json:"sentAt"`

	TodoObjectID bson.ObjectId `json:"todoObjectID" bson:",omitempty"`
}

//Create a noptification out of a date
func (surveyCatalogNotification *SurveyCatalogNotification) Create(notificationDate time.Time) {
	if surveyCatalogNotification.ID.Hex() == "" || !surveyCatalogNotification.ID.Valid() {
		surveyCatalogNotification.ID = bson.NewObjectId()
	}

	surveyCatalogNotification.IsSent = false
	surveyCatalogNotification.Date = notificationDate
}

//SurveyCatalogPlan is the plan for when and how a survey will have to be conducted
type SurveyCatalogPlan struct {
	Survey        Survey                      `json:"survey"`
	StartDate     time.Time                   `json:"startDate"`
	EndDate       time.Time                   `json:"endDate"`
	Type          string                      `json:"type"`
	Notifications []SurveyCatalogNotification `json:"notifications"`
}

//SurveyCatalog model
type SurveyCatalog struct {
	BaseModel `bson:",inline"`

	Name    string  `json:"name" valid:"required"`
	Company Company `json:"company" valid:"required"`

	BaseSurveyPlan       SurveyCatalogPlan   `json:"baseSurveyPlan,omitempty" bson:",omitempty"`
	RecurringSurveyPlans []SurveyCatalogPlan `json:"recurringSurveyPlans,omitempty" bson:",omitempty"`
	EndSurveyPlan        SurveyCatalogPlan   `json:"endSurveyPlan,omitempty" bson:",omitempty"`

	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`

	EMailTemplateID bson.ObjectId `json:"eMailTemplateId" bson:",omitempty"`

	TicketTodos []SurveyTicketTodo `json:"-" bson:",omitempty"`

	RawCatalogSettings SurveyCatalogAddSettingsRequestModel `json:"-" bson:",omitempty"`

	IsDraft  bool `json:"isDraft"`
	IsActive bool `json:"isActive"`
	IsDone   bool `json:"isDone"`

	IsEvaluated  bool   `json:"-"`
	EvaluationID string `json:"-" bson:",omitempty"`
}

func (surveyCatalog SurveyCatalog) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model":      "SurveyCatalog",
		"collection": CollectionSurveyCatalogProperty,
	})
}

//Validate the given survey catalog
func (surveyCatalog *SurveyCatalog) Validate() error {
	_, err := govalidator.ValidateStruct(surveyCatalog)
	if err != nil {
		return err
	}
	return nil
}

func getModuleIDsFromSurvey(survey Survey) ([]bson.ObjectId, error) {
	moduleIDs := make([]bson.ObjectId, 0)
	for _, module := range survey.Modules {
		moduleIDs = append(moduleIDs, module.ID)
	}
	return moduleIDs, nil
}

//GetSurveyModulesIDs returns the selected modules as IDs
//surveyType: 0 = "BaseSurvey", 1 = "RecurringSurvey", 2 = "EndSurvey"
func (surveyCatalog *SurveyCatalog) GetSurveyModulesIDs(catalogID string, surveyType int) ([]bson.ObjectId, error) {
	if surveyType == 0 {
		return getModuleIDsFromSurvey(surveyCatalog.BaseSurveyPlan.Survey)
	}
	if surveyType == 1 {
		if len(surveyCatalog.RecurringSurveyPlans) > 0 {
			return getModuleIDsFromSurvey(surveyCatalog.RecurringSurveyPlans[0].Survey)
		}
		emptyObjectIDArray := make([]bson.ObjectId, 0)
		return emptyObjectIDArray, nil
	}
	if surveyType == 2 {
		return getModuleIDsFromSurvey(surveyCatalog.EndSurveyPlan.Survey)
	}

	return nil, ErrRecordNotFound
}

//ValidateAndAddSurveyModules validates the provided module ID and checks their existence
//It also creates the survey and adds assignes it to the catalog
//surveyType: 0 = "BaseSurvey", 1 = "RecurringSurvey", 2 = "EndSurvey"
func (surveyCatalog *SurveyCatalog) ValidateAndAddSurveyModules(catalogID string, surveyType int, requestBody io.ReadCloser) error {
	if err := surveyCatalog.FindByID(catalogID); err != nil {
		return err
	}

	if !surveyCatalog.IsDraft {
		return ErrSurveyCatalogAlreadyPublished
	}

	decoder := json.NewDecoder(requestBody)
	var requestData []bson.ObjectId
	if err := decoder.Decode(&requestData); err != nil {
		if err != io.EOF {
			return err
		}
		return ErrEmptyRequestBody
	}

	var survey Survey
	if err := survey.Create(requestData); err != nil {
		return err
	}

	surveyPlan := SurveyCatalogPlan{
		Survey: survey,
	}

	if surveyType == 0 {
		surveyPlan.Type = SurveyTodoTypeBase
		surveyCatalog.BaseSurveyPlan = surveyPlan
	}
	if surveyType == 1 {
		surveyPlan.Type = SurveyTodoTypeRecurring
		surveyCatalog.RecurringSurveyPlans = []SurveyCatalogPlan{surveyPlan}
	}
	if surveyType == 2 {
		surveyPlan.Type = SurveyTodoTypeFinish
		surveyCatalog.EndSurveyPlan = surveyPlan
	}

	return surveyCatalog.Save()
}

//validateNotificationDates validates the notification dates and checks if they are in the timespan of the
//actual survey and are each after one another
func validateNotificationDates(startDate time.Time, endDate time.Time, notifications []SurveyCatalogNotification) error {
	for notificationTimeIdx, notification := range notifications {
		if notification.Date.Before(startDate) {
			return ErrSurveyCatalogNotificationBeforeStart
		}
		if notification.Date.After(endDate) {
			return ErrSurveyCatalogNotificationAfterEnd
		}

		if len(notifications) > 1 {
			notificationTimeBeforeIndex := notificationTimeIdx - 1
			if notificationTimeBeforeIndex > 0 && notificationTimeBeforeIndex <= len(notifications) {
				if !notification.Date.After(notifications[notificationTimeBeforeIndex].Date) {
					return ErrSurveyCatalogNotificationBeforeLast
				}
			}
		}
	}
	return nil
}

//ValidateAndAddRecurringSettings validates and adds each repeat of a specific survey to the as single plan
func (surveyCatalog *SurveyCatalog) ValidateAndAddRecurringSettings(recurringSurveySettings SurveyCatalogAddSettingsRecurringSettingsModel) error {
	if !surveyCatalog.IsDraft {
		return ErrSurveyCatalogAlreadyPublished
	}

	// Generate reference dates
	now := time.Now()

	// Recurring date have to be after the endDate of the base survey
	if !recurringSurveySettings.StartDate.After(surveyCatalog.BaseSurveyPlan.EndDate) {
		return ErrSurveyCatalogStartDateBeforeEndDate
	}

	if !recurringSurveySettings.StartDate.Before(recurringSurveySettings.EndDate) {
		return ErrSurveyCatalogStartDateAfterEndDate
	}
	if !recurringSurveySettings.StartDate.After(now) {
		return ErrSurveyCatalogStartDateInPast
	}

	parsedNotifications := make([]SurveyCatalogNotification, 0)
	for _, notificationTime := range recurringSurveySettings.Notifications {
		var notification SurveyCatalogNotification
		notification.Create(notificationTime)
		parsedNotifications = append(parsedNotifications, notification)
	}

	if err := validateNotificationDates(recurringSurveySettings.StartDate, recurringSurveySettings.EndDate, parsedNotifications); err != nil {
		return err
	}

	if len(surveyCatalog.RecurringSurveyPlans) < 1 {
		return ErrSurveyCatalogAlreadyPublished
	}

	basePlan := surveyCatalog.RecurringSurveyPlans[0]
	basePlan.StartDate = recurringSurveySettings.StartDate
	basePlan.EndDate = recurringSurveySettings.EndDate
	basePlan.Notifications = parsedNotifications

	recurringSurveyPlans := make([]SurveyCatalogPlan, 0)
	recurringSurveyPlans = append(recurringSurveyPlans, basePlan)

	for i := 1; i <= recurringSurveySettings.Repeats; i++ {
		weekAddition := time.Hour * 24 * 7 * time.Duration(i)
		repeatStartDate := basePlan.StartDate.Add(weekAddition)
		repeatEndDate := basePlan.EndDate.Add(weekAddition)

		if !repeatStartDate.Before(repeatEndDate) {
			return ErrSurveyCatalogStartDateAfterEndDate
		}

		lastRecurringPlanIndex := i - 1
		if lastRecurringPlanIndex > 0 && lastRecurringPlanIndex <= len(recurringSurveyPlans) {
			lastRecurringPlan := recurringSurveyPlans[lastRecurringPlanIndex]
			if !repeatStartDate.After(lastRecurringPlan.EndDate) {
				return ErrSurveyCatalogStartDateBeforeEndDate
			}
		}

		recurringNotifications := make([]SurveyCatalogNotification, 0)
		for _, notification := range basePlan.Notifications {
			notification.Date = notification.Date.Add(weekAddition)
			recurringNotifications = append(recurringNotifications, notification)
		}

		if err := validateNotificationDates(repeatStartDate, repeatEndDate, recurringNotifications); err != nil {
			return err
		}

		basePlan.StartDate = repeatStartDate
		basePlan.EndDate = repeatEndDate
		basePlan.Type = SurveyTodoTypeRecurring
		basePlan.Notifications = recurringNotifications
		recurringSurveyPlans = append(recurringSurveyPlans, basePlan)
	}

	surveyCatalog.RecurringSurveyPlans = recurringSurveyPlans

	return nil
}

//AddSurveySettings validates and adds the settings to a catalog
func (surveyCatalog *SurveyCatalog) AddSurveySettings(requestSettings SurveyCatalogAddSettingsRequestModel) error {
	if !surveyCatalog.ID.Valid() {
		return ErrInvalidObjectID
	}

	if !surveyCatalog.IsDraft {
		return ErrSurveyCatalogAlreadyPublished
	}

	// Validate general request for complete data
	if _, err := govalidator.ValidateStruct(&requestSettings); err != nil {
		return err
	}

	// Generate reference dates
	now := time.Now()

	// Validate base survey settings
	baseSurveySettings := requestSettings.Base

	if !baseSurveySettings.StartDate.Before(baseSurveySettings.EndDate) {
		return ErrSurveyCatalogStartDateAfterEndDate
	}
	if !baseSurveySettings.StartDate.After(now) {
		return ErrSurveyCatalogStartDateInPast
	}

	parsedBaseNotifications := make([]SurveyCatalogNotification, 0)
	for _, notificationTime := range baseSurveySettings.Notifications {
		var notification SurveyCatalogNotification
		notification.Create(notificationTime)
		parsedBaseNotifications = append(parsedBaseNotifications, notification)
	}

	if err := validateNotificationDates(baseSurveySettings.StartDate, baseSurveySettings.EndDate, parsedBaseNotifications); err != nil {
		return err
	}

	// Assign date to actual survey plan
	surveyCatalog.BaseSurveyPlan.StartDate = baseSurveySettings.StartDate
	surveyCatalog.BaseSurveyPlan.EndDate = baseSurveySettings.EndDate
	surveyCatalog.BaseSurveyPlan.Notifications = parsedBaseNotifications
	surveyCatalog.BaseSurveyPlan.Type = SurveyTodoTypeBase

	if len(surveyCatalog.RecurringSurveyPlans) > 0 {
		// Validate recurring survey settings
		recurringSurveySettings := requestSettings.Recurring

		if err := surveyCatalog.ValidateAndAddRecurringSettings(recurringSurveySettings); err != nil {
			return err
		}

		// Recurring date have to be after the endDate of the base survey
		if recurringSurveySettings.StartDate.Before(baseSurveySettings.EndDate) {
			return ErrSurveyCatalogStartDateBeforeEndDate
		}
	}

	// Validate end survey settings
	endSurveySettings := requestSettings.End

	if !endSurveySettings.StartDate.Before(endSurveySettings.EndDate) {
		return ErrSurveyCatalogStartDateAfterEndDate
	}
	if !endSurveySettings.StartDate.After(now) {
		return ErrSurveyCatalogStartDateInPast
	}

	parsedEndNotifications := make([]SurveyCatalogNotification, 0)
	for _, notificationTime := range endSurveySettings.Notifications {
		var notification SurveyCatalogNotification
		notification.Create(notificationTime)
		parsedEndNotifications = append(parsedEndNotifications, notification)
	}

	if err := validateNotificationDates(endSurveySettings.StartDate, endSurveySettings.EndDate, parsedEndNotifications); err != nil {
		return err
	}

	lastRecurringSurvey := surveyCatalog.BaseSurveyPlan
	if len(surveyCatalog.RecurringSurveyPlans) > 0 {
		lastRecurringSurvey = surveyCatalog.RecurringSurveyPlans[len(surveyCatalog.RecurringSurveyPlans)-1]
	}

	// End survey date have to be after the endDate of the last recurring survey
	if endSurveySettings.StartDate.Before(lastRecurringSurvey.EndDate) {
		return ErrSurveyCatalogStartDateBeforeEndDate
	}

	// Assign date to actual survey plan
	surveyCatalog.EndSurveyPlan.StartDate = endSurveySettings.StartDate
	surveyCatalog.EndSurveyPlan.EndDate = endSurveySettings.EndDate
	surveyCatalog.EndSurveyPlan.Notifications = parsedEndNotifications
	surveyCatalog.EndSurveyPlan.Type = SurveyTodoTypeFinish

	surveyCatalog.StartDate = surveyCatalog.BaseSurveyPlan.StartDate
	surveyCatalog.EndDate = surveyCatalog.EndSurveyPlan.EndDate

	surveyCatalog.RawCatalogSettings = requestSettings

	return surveyCatalog.Save()
}

//AddTodoObjectIDsToNotifications adds the generated todods back to the notifications
func (surveyCatalog *SurveyCatalog) AddTodoObjectIDsToNotifications(todos []SurveyTicketTodo) error {
	if !surveyCatalog.ID.Valid() {
		return ErrInvalidObjectID
	}

	recurringTodoIdx := 0
	for _, todo := range todos {
		if todo.Type == SurveyTodoTypeBase {
			for notificationIdx := range surveyCatalog.BaseSurveyPlan.Notifications {
				surveyCatalog.BaseSurveyPlan.Notifications[notificationIdx].TodoObjectID = todo.ID
			}
		}
		if todo.Type == SurveyTodoTypeRecurring {
			for notificationIdx := range surveyCatalog.RecurringSurveyPlans[recurringTodoIdx].Notifications {
				surveyCatalog.RecurringSurveyPlans[recurringTodoIdx].Notifications[notificationIdx].TodoObjectID = todo.ID
			}
			recurringTodoIdx++
		}
		if todo.Type == SurveyTodoTypeFinish {
			for notificationIdx := range surveyCatalog.EndSurveyPlan.Notifications {
				surveyCatalog.EndSurveyPlan.Notifications[notificationIdx].TodoObjectID = todo.ID
			}
		}
	}

	return surveyCatalog.Save()
}

//Publish publishes the given aurvey
func (surveyCatalog *SurveyCatalog) Publish() error {
	if !surveyCatalog.ID.Valid() {
		return ErrInvalidObjectID
	}
	if !surveyCatalog.IsDraft {
		return ErrSurveyCatalogAlreadyPublished
	}

	if err := surveyCatalog.GenerateTickets(); err != nil {
		return err
	}

	surveyCatalog.IsDraft = false

	return surveyCatalog.Save()
}

//GenerateTickets creates the tickets for all participants
func (surveyCatalog *SurveyCatalog) GenerateTickets() error {
	if !surveyCatalog.ID.Valid() {
		return ErrInvalidObjectID
	}

	company := surveyCatalog.Company
	if !company.ID.Valid() {
		return ErrInvalidObjectID
	}

	// generate todods
	ticketTodos := make([]SurveyTicketTodo, 0)

	var baseSurveyTicketTodo SurveyTicketTodo
	baseSurveyTicketTodo.Create(surveyCatalog.BaseSurveyPlan)
	ticketTodos = append(ticketTodos, baseSurveyTicketTodo)

	for _, recurringSurveyPlan := range surveyCatalog.RecurringSurveyPlans {
		var recurringSurveyTicketTodo SurveyTicketTodo
		recurringSurveyTicketTodo.Create(recurringSurveyPlan)
		ticketTodos = append(ticketTodos, recurringSurveyTicketTodo)
	}

	var endSurveyTicketTodo SurveyTicketTodo
	endSurveyTicketTodo.Create(surveyCatalog.EndSurveyPlan)
	ticketTodos = append(ticketTodos, endSurveyTicketTodo)

	surveyCatalog.TicketTodos = ticketTodos

	// calculate dates (start and end)
	surveyTicketStartDate := ticketTodos[0].StartDate
	surveyTicketEndDate := ticketTodos[len(ticketTodos)-1].DueDate

	surveyCatalog.StartDate = surveyTicketStartDate
	surveyCatalog.EndDate = surveyTicketEndDate

	if err := surveyCatalog.Save(); err != nil {
		return err
	}

	amqpConn, err := GetAMQPClient()
	if err != nil {
		return err
	}
	defer amqpConn.Close()

	amqpSurveyTicketCreateItem := AMQPSurveyTicketCreateItem{}
	amqpChannel, err := amqpSurveyTicketCreateItem.DeclareChannel(amqpConn)
	if err != nil {
		return err
	}
	defer amqpChannel.Close()

	// create worker items for ticket creation
	participants, err := company.FindParticipants()
	if err != nil {
		return err
	}

	for _, participant := range participants {
		amqpSurveyTicketCreateItem := AMQPSurveyTicketCreateItem{
			SurveyCatalogObjectID: surveyCatalog.ID,
			ParticipantObjectID:   participant.ID,
		}
		if err := amqpSurveyTicketCreateItem.PublishItem(amqpChannel); err != nil {
			return err
		}
	}

	return surveyCatalog.AddTodoObjectIDsToNotifications(ticketTodos)
}

//GetLightTickets returns just enough information to send out notifications
//it omits data such as the todods
func (surveyCatalog *SurveyCatalog) GetLightTickets() ([]SurveyTicket, error) {
	if !surveyCatalog.ID.Valid() {
		return nil, ErrInvalidObjectID
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return nil, err
	}

	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	surveyCatalogIDQuery := bson.M{
		"surveycatalogid": surveyCatalog.ID,
	}

	surveyTickets := make([]SurveyTicket, 0)
	if err := ticketsCol.Find(surveyCatalogIDQuery).Select(bson.M{"todos.questions": 0}).All(&surveyTickets); err != nil {
		return nil, err
	}

	return surveyTickets, nil
}

// get light tickets for support
func (surveyCatalog *SurveyCatalog) GetSupportTickets() ([]SurveyTicket, error) {
	if !surveyCatalog.ID.Valid() {
		return nil, ErrInvalidObjectID
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return nil, err
	}

	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	surveyCatalogIDQuery := bson.M{
		"surveycatalogid": surveyCatalog.ID,
	}

	surveyTickets := make([]SurveyTicket, 0)
	if err := ticketsCol.Find(surveyCatalogIDQuery).Select(bson.M{"todos.questions": 0, "authcode": 0}).All(&surveyTickets); err != nil {
		return nil, err
	}

	return surveyTickets, nil
}

//FindTickets returns all tickets for a given survey
func (surveyCatalog *SurveyCatalog) FindTickets() ([]SurveyTicket, error) {
	if !surveyCatalog.ID.Valid() {
		return nil, ErrInvalidObjectID
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return nil, err
	}

	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	surveyCatalogIDQuery := bson.M{
		"surveycatalogid": surveyCatalog.ID,
	}

	surveyTickets := make([]SurveyTicket, 0)
	if err := ticketsCol.Find(surveyCatalogIDQuery).All(&surveyTickets); err != nil {
		return nil, err
	}

	return surveyTickets, nil
}

//Save or updates the given survey catalog
func (surveyCatalog *SurveyCatalog) Save() error {
	if err := surveyCatalog.Validate(); err != nil {
		return err
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	rolesCol := dataStore.GetCollection(CollectionSurveyCatalogProperty)
	defer dataStore.Close()

	surveyCatalog.SetID()
	surveyCatalog.UpdatedAt = time.Now()

	_, err := rolesCol.UpsertId(surveyCatalog.ID, bson.M{"$set": surveyCatalog})
	return err
}

//FindAll returns all survey catalogs
func (surveyCatalog *SurveyCatalog) FindAll() ([]SurveyCatalog, error) {
	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return nil, dataStoreErr
	}
	surveyModuleCol := dataStore.GetCollection(CollectionSurveyCatalogProperty)
	defer dataStore.Close()

	surveyCatalogs := make([]SurveyCatalog, 0)

	excludeDraftsQuery := bson.M{"isdraft": false}

	err := surveyModuleCol.Find(nil).All(&surveyCatalogs)
	if err != nil {
		surveyCatalog.getLogger().WithFields(logrus.Fields{
			"query": excludeDraftsQuery,
		}).Debug(err)
		return nil, ErrRecordsFetch
	}

	return surveyCatalogs, nil
}

//FindByID returns a survey catalog
func (surveyCatalog *SurveyCatalog) FindByID(id string) error {
	if isObjectID := bson.IsObjectIdHex(id); !isObjectID {
		return ErrInvalidObjectID
	}
	return surveyCatalog.FindByObjectID(bson.ObjectIdHex(id))
}

//FindByObjectID returns a survey catalof
func (surveyCatalog *SurveyCatalog) FindByObjectID(surveyCatalogID bson.ObjectId) error {
	if isValid := surveyCatalogID.Valid(); !isValid {
		return ErrInvalidObjectID
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	companiesCol := dataStore.GetCollection(CollectionSurveyCatalogProperty)
	defer dataStore.Close()

	err := companiesCol.FindId(surveyCatalogID).One(&surveyCatalog)
	if err != nil {
		surveyCatalog.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("FindId(%s))", surveyCatalogID),
			"error": err.Error(),
		}).Debug(err)
		return err
	}

	return nil
}

//EnableAllTickets when survey is activated
func (surveyCatalog *SurveyCatalog) EnableAllTickets() error {
	if !surveyCatalog.ID.Valid() {
		return ErrInvalidObjectID
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}

	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	setEnableAllTicketsChange := bson.M{
		"$set": bson.M{
			"isactive":  true,
			"updatedAt": time.Now(),
		},
	}

	surveyTicketSurveyCatalogIDQuery := bson.M{
		"surveycatalogid": surveyCatalog.ID,
		"isactive":        false,
	}

	if _, err := ticketsCol.UpdateAll(surveyTicketSurveyCatalogIDQuery, setEnableAllTicketsChange); err != nil {
		surveyCatalog.getLogger().Info(err)
		return ErrRecordNotDeleteable
	}

	return nil
}

//DisableAllTickets when survey is done
func (surveyCatalog *SurveyCatalog) DisableAllTickets() error {
	if !surveyCatalog.ID.Valid() {
		return ErrInvalidObjectID
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}

	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	setDisableAllTicketsChange := bson.M{
		"$set": bson.M{
			"isactive":  false,
			"updatedAt": time.Now(),
		},
	}

	surveyTicketSurveyCatalogIDQuery := bson.M{
		"surveycatalogid": surveyCatalog.ID,
		"isactive":        true,
	}

	if _, err := ticketsCol.UpdateAll(surveyTicketSurveyCatalogIDQuery, setDisableAllTicketsChange); err != nil {
		surveyCatalog.getLogger().Info(err)
		return ErrRecordNotDeleteable
	}

	return nil
}

//AnonymizeAndDisableTickets removes all data from the tickets
func (surveyCatalog *SurveyCatalog) AnonymizeAndDisableTickets() error {
	if !surveyCatalog.ID.Valid() {
		return ErrInvalidObjectID
	}

	if !surveyCatalog.IsDone {
		return ErrSurveyCatalogStillActive
	}

	dataStore, err := CreateDataStore()
	if err != nil {
		return err
	}

	ticketsCol := dataStore.GetCollection(CollectionTicketsProperty)
	defer dataStore.Close()

	removeAllPersonalDataChange := bson.M{
		"$set": bson.M{
			"isactive":  false,
			"updatedAt": time.Now(),
		},
		"$unset": bson.M{
			"participant":       1,
			"sentnotifications": 1,
			"authcode":          1,
		},
	}

	surveyTicketSurveyCatalogIDQuery := bson.M{
		"surveycatalogid": surveyCatalog.ID,
	}

	if _, err := ticketsCol.UpdateAll(surveyTicketSurveyCatalogIDQuery, removeAllPersonalDataChange); err != nil {
		fmt.Println(err)
		surveyCatalog.getLogger().Info(err)
		return err
	}

	return nil
}

//Delete survey catalog
func (surveyCatalog *SurveyCatalog) Delete() error {
	if !surveyCatalog.ID.Valid() {
		return ErrRecordNotFound
	}

	if surveyCatalog.IsActive || surveyCatalog.IsDone {
		return ErrSurveyCatalogStillActive
	}

	dataStore, dataStoreErr := CreateDataStore()
	if dataStoreErr != nil {
		return dataStoreErr
	}
	surveyCatalogCol := dataStore.GetCollection(CollectionSurveyCatalogProperty)
	defer dataStore.Close()

	err := surveyCatalogCol.RemoveId(surveyCatalog.ID)
	if err != nil {
		surveyCatalog.getLogger().WithFields(logrus.Fields{
			"query": fmt.Sprintf("RemoveId(%s))", surveyCatalog.ID),
			"error": err.Error(),
		}).Error(err)
		return ErrRecordNotDeleteable
	}

	return nil
}
