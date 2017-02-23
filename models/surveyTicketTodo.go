package models

import "time"

const (
	//SurveyTodoStateLocked locked survey
	SurveyTodoStateLocked = "locked"
	//SurveyTodoStateActive active survey
	SurveyTodoStateActive = "active"
	//SurveyTodoStateDone done survey
	SurveyTodoStateDone = "done"
	//SurveyTodoStateExpired expired survey
	SurveyTodoStateExpired = "expired"
)

const (
	//SurveyTodoTypeBase base survey
	SurveyTodoTypeBase = "base"
	//SurveyTodoTypeRecurring weekly survey
	SurveyTodoTypeRecurring = "recurring"
	//SurveyTodoTypeFinish end survey
	SurveyTodoTypeFinish = "end"
)

//SurveyTicketTodo model
type SurveyTicketTodo struct {
	BaseModel `bson:",inline"`

	Type string `json:"type" valid:"required"`

	Status   string  `json:"status"`
	Progress float32 `json:"progress"`

	StartDate time.Time `json:"startDate"`
	DueDate   time.Time `json:"dueDate"`

	EstimatedDuration float32 `json:"estimatedDuration"`
	MeasuredDuration  float32 `json:"measuredDuration,omitempty"`

	IsEnriched bool      `json:"isEnriched"`
	EnrichedAt time.Time `json:"enrichedAt"`

	SubmittedAt           time.Time `json:"submittedAt,omitempty" bson:",omitempty"`
	ApplicationIdentifier string    `json:"submittedClientIdentifier,omitempty" bson:",omitempty"`

	Questions []Question `json:"questions"`
}

//Create a new todo item out of modules
func (surveyTicketTodo *SurveyTicketTodo) Create(surveyPlan SurveyCatalogPlan) error {
	surveyTicketTodo.SetID()
	surveyTicketTodo.UpdatedAt = time.Now()
	surveyTicketTodo.Status = SurveyTodoStateLocked
	surveyTicketTodo.Progress = 0

	for _, module := range surveyPlan.Survey.Modules {
		surveyTicketTodo.EstimatedDuration = module.EstimatedDuration
		for _, question := range module.Questions {
			surveyTicketTodo.Questions = append(surveyTicketTodo.Questions, question)
		}
	}

	surveyTicketTodo.StartDate = surveyPlan.StartDate
	surveyTicketTodo.DueDate = surveyPlan.EndDate
	surveyTicketTodo.Type = surveyPlan.Type

	return nil
}
