package controllers

import (
	"net/http"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/unrolled/render"
)

//VersionCtrl is the controller for /version
type VersionCtrl struct{}

func (versionCtrl VersionCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "VersionCtrl",
	})
}

//List shows current version
func (versionCtrl VersionCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	resMap := make(map[string]string)
	resMap["version"] = os.Getenv("VERSION")

	r.JSON(res, 200, resMap)
	return
}
