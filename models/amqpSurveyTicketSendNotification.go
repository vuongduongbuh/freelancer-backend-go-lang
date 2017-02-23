package models

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/mgo.v2/bson"
)

var (
	//AMQPSurveyTicketSendNotificationQueueName queue name
	AMQPSurveyTicketSendNotificationQueueName = "notification_send"
)

//AMQPSurveyTicketSendNotification is used as AMQP model
type AMQPSurveyTicketSendNotification struct {
	SurveyTicketObjectID     bson.ObjectId `valid:"required" json:"s"`
	SurveyTicketTodoObjectID bson.ObjectId `valid:"required" json:"t"`
	NotificationObjectID     bson.ObjectId `valid:"required" json:"n"`
	EMailTemplateObjectID    bson.ObjectId `json:"e,omitempty"`
}

//getLogger
func (amqpSurveyTicketSendNotification *AMQPSurveyTicketSendNotification) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "AMQPSurveyTicketSendNotification",
	})
}

//ConvertToBytes converts message to bytes for the queue
func (amqpSurveyTicketSendNotification *AMQPSurveyTicketSendNotification) ConvertToBytes() ([]byte, error) {
	body, err := json.Marshal(amqpSurveyTicketSendNotification)
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

//ConvertFromBytes gets the message and parse it
func (amqpSurveyTicketSendNotification *AMQPSurveyTicketSendNotification) ConvertFromBytes(msg []byte) error {
	return json.Unmarshal(msg, &amqpSurveyTicketSendNotification)
}

//DeclareChannel adds item to queue
func (amqpSurveyTicketSendNotification *AMQPSurveyTicketSendNotification) DeclareChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		AMQPSurveyTicketSendNotificationQueueName, // name
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
func (amqpSurveyTicketSendNotification *AMQPSurveyTicketSendNotification) Publish(ch *amqp.Channel) error {
	body, err := amqpSurveyTicketSendNotification.ConvertToBytes()
	if err != nil {
		return err
	}

	return ch.Publish(
		"", // exchange
		AMQPSurveyTicketSendNotificationQueueName, // routing key
		true, // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}
