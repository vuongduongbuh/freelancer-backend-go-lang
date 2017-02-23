package models

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/mgo.v2/bson"
)

var (
	//AMQPSurveyTicketExchangeName tickets exchange
	AMQPSurveyTicketExchangeName = "tickets_exchange"
	//AMQPSurveyTicketCreateQueueName queue name
	AMQPSurveyTicketCreateQueueName = "ticket_create"
)

//AMQPSurveyTicketCreateItem is used as AMQP model
type AMQPSurveyTicketCreateItem struct {
	SurveyCatalogObjectID bson.ObjectId `valid:"required" json:"soi"`
	ParticipantObjectID   bson.ObjectId `valid:"required"  json:"poi"`
}

//getLogger
func (amqpSurveyTicketCreateItem *AMQPSurveyTicketCreateItem) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "AMQPSurveyTicketCreatetem",
	})
}

//ConvertToBytes converts message to bytes for the queue
func (amqpSurveyTicketCreateItem *AMQPSurveyTicketCreateItem) ConvertToBytes() ([]byte, error) {
	body, err := json.Marshal(amqpSurveyTicketCreateItem)
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

//ConvertFromBytes gets the message and parse it
func (amqpSurveyTicketCreateItem *AMQPSurveyTicketCreateItem) ConvertFromBytes(msg []byte) error {
	return json.Unmarshal(msg, &amqpSurveyTicketCreateItem)
}

//DeclareChannel adds item to queue
func (amqpSurveyTicketCreateItem *AMQPSurveyTicketCreateItem) DeclareChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		AMQPSurveyTicketCreateQueueName, // name
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

//PublishItem adds item to queue
func (amqpSurveyTicketCreateItem *AMQPSurveyTicketCreateItem) PublishItem(ch *amqp.Channel) error {
	body, err := amqpSurveyTicketCreateItem.ConvertToBytes()
	if err != nil {
		return err
	}

	return ch.Publish(
		"", // exchange
		AMQPSurveyTicketCreateQueueName, // routing key
		true, // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}
