package models

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/mgo.v2/bson"
)

var (
	//AMQPSurveyTicketSendInvitationEMailQueueName queue name
	AMQPSurveyTicketSendInvitationEMailQueueName = "notification_sendinvitation"
)

//AMQPSurveyTicketSendInvitationEMail is used as AMQP model
type AMQPSurveyTicketSendInvitationEMail struct {
	SurveyTicketObjectID  bson.ObjectId `valid:"required" json:"s"`
	EMailTemplateObjectID bson.ObjectId `valid:"required" json:"e"`
	Resend                bool          `json:"f"`
}

//getLogger
func (amqpSurveyTicketSendInvitationEMail *AMQPSurveyTicketSendInvitationEMail) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "AMQPSurveyTicketSendInvitationEMail",
	})
}

//ConvertToBytes converts message to bytes for the queue
func (amqpSurveyTicketSendInvitationEMail *AMQPSurveyTicketSendInvitationEMail) ConvertToBytes() ([]byte, error) {
	body, err := json.Marshal(amqpSurveyTicketSendInvitationEMail)
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

//ConvertFromBytes gets the message and parse it
func (amqpSurveyTicketSendInvitationEMail *AMQPSurveyTicketSendInvitationEMail) ConvertFromBytes(msg []byte) error {
	return json.Unmarshal(msg, &amqpSurveyTicketSendInvitationEMail)
}

//DeclareChannel adds item to queue
func (amqpSurveyTicketSendInvitationEMail *AMQPSurveyTicketSendInvitationEMail) DeclareChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		AMQPSurveyTicketSendInvitationEMailQueueName, // name
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
func (amqpSurveyTicketSendInvitationEMail *AMQPSurveyTicketSendInvitationEMail) Publish(ch *amqp.Channel) error {
	body, err := amqpSurveyTicketSendInvitationEMail.ConvertToBytes()
	if err != nil {
		return err
	}

	return ch.Publish(
		"", // exchange
		AMQPSurveyTicketSendInvitationEMailQueueName, // routing key
		true, // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}
