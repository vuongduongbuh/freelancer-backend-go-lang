package controllers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/unrolled/render"
)

//WorkbenchCtrl is the controller for /companies
type WorkbenchCtrl struct{}

func (workbenchCtrl WorkbenchCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "WorkbenchCtrl",
	})
}

func generateRandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

//Test all users
func (workbenchCtrl WorkbenchCtrl) Test(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	r.Text(res, 201, "surveyTicket")
	return
}
