package controllers

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/unrolled/render"
)

//AuthCtrl /auth
type AuthCtrl struct{}

func (authCtrl AuthCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "AuthCtrl",
	})
}

//Check current auth token
func (authCtrl AuthCtrl) Check(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	r.Text(res, 204, "")
}
