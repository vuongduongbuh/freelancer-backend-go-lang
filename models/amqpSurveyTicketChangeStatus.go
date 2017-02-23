package models

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/mgo.v2/bson"
)

var (
	//AMQPSurveyTicketChangeStateQueueName queue name
	AMQPSurveyTicketChangeStateQueueName = "ticket_mutatestate"
)

//AMQPSurveyTicketTodoState is used for changed states
type AMQPSurveyTicketTodoState struct {
	SurveyTicketTodoObjectID bson.ObjectId `valid:"required" json:"t"`
	State                    string        `valid:"required" json:"s"`
}

//AMQPSurveyTicketChangeStateItem is used as AMQP model
type AMQPSurveyTicketChangeStateItem struct {
	SurveyTicketObjectID       bson.ObjectId               `valid:"required" json:"so"`
	SurveyTicketShouldBeDelete bool                        `json:"sd"`
	SurveyTicketTodosChanged   bool                        `json:"ts"`
	SurveyTicketTodoStates     []AMQPSurveyTicketTodoState `json:"ot,omitempty"`
}

//getLogger
func (amqpSurveyTicketChangeStateItem *AMQPSurveyTicketChangeStateItem) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "AMQPSurveyTicketChangeStateItem",
	})
}

//ConvertToBytes converts message to bytes for the queue
func (amqpSurveyTicketChangeStateItem *AMQPSurveyTicketChangeStateItem) ConvertToBytes() ([]byte, error) {
	body, err := json.Marshal(amqpSurveyTicketChangeStateItem)
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

//ConvertFromBytes gets the message and parse it
func (amqpSurveyTicketChangeStateItem *AMQPSurveyTicketChangeStateItem) ConvertFromBytes(msg []byte) error {
	return json.Unmarshal(msg, &amqpSurveyTicketChangeStateItem)
}

//DeclareChannel adds item to queue
func (amqpSurveyTicketChangeStateItem *AMQPSurveyTicketChangeStateItem) DeclareChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		AMQPSurveyTicketChangeStateQueueName, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

//Publish adds item to queue
func (amqpSurveyTicketChangeStateItem *AMQPSurveyTicketChangeStateItem) Publish(ch *amqp.Channel) error {
	body, err := amqpSurveyTicketChangeStateItem.ConvertToBytes()
	if err != nil {
		return err
	}

	return ch.Publish(
		"", // exchange
		AMQPSurveyTicketChangeStateQueueName, // routing key
		true, // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

//GetTodoState returns todo state
func (amqpSurveyTicketChangeStateItem *AMQPSurveyTicketChangeStateItem) GetTodoState(todoID bson.ObjectId) string {
	for _, todoState := range amqpSurveyTicketChangeStateItem.SurveyTicketTodoStates {
		if todoState.SurveyTicketTodoObjectID != todoID {
			continue
		}
		return todoState.State
	}
	return "locked"
}
