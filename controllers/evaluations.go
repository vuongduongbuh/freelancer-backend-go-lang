package controllers

import "github.com/Sirupsen/logrus"

//EvaluationsCtrl is the controller for /evaluations
type EvaluationsCtrl struct{}

func (evaluationsCtrl EvaluationsCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "EvaluationsCtrl",
	})
}
