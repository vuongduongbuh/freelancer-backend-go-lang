package controllers

import (
	"encoding/json"
	"net/http"

	"gopkg.in/mgo.v2/bson"

	"gitlab.com/slsurvey/slsurvey-srv/helpers"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//UsersCtrl is the controller for /users
type UsersCtrl struct{}

func (usersCtrl UsersCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "UsersCtrl",
	})
}

//List all users
func (usersCtrl UsersCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	id := vars["id"]

	var user models.User

	if id != "" {
		if err := user.FindByID(id); err != nil {
			jsonErr := helpers.GenerateJSONError(err, req.Header)
			r.JSON(res, jsonErr.StatusCode, jsonErr)
			return
		}

		r.JSON(res, 200, user)
		return
	}

	users, err := user.FindAll()
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, users)
	return
}

type userCreateRequestModel struct {
	EMail           string        `json:"email"`
	Password        string        `json:"password"`
	CompanyObjectID bson.ObjectId `json:"companyId"`
}

//Create user
func (usersCtrl UsersCtrl) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var requestData userCreateRequestModel
	if err := decoder.Decode(&requestData); err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrBadRequest, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	var user models.User
	if err := user.Create(requestData.EMail, requestData.Password, requestData.CompanyObjectID); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	if err := user.FindByID(user.ID); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.JSON(res, 200, user)
	return
}

//Update user
func (usersCtrl UsersCtrl) Update(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	userID := vars["id"]

	if userID == "" {
		jsonErr := helpers.GenerateJSONError(helpers.ErrBadRequest, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	var requestData models.UserUpdateRequestModel

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&requestData); err != nil {
		jsonErr := helpers.GenerateJSONError(helpers.ErrBadRequest, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	var user models.User
	if err := user.FindByID(userID); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	if err := user.Update(requestData); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.JSON(res, 200, user)
}

//ToggleBlock user
func (usersCtrl UsersCtrl) ToggleBlock(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	userID := vars["id"]

	if userID == "" {
		jsonErr := helpers.GenerateJSONError(helpers.ErrBadRequest, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	var user models.User
	if err := user.FindByID(userID); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	if err := user.ToogleBlock(); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.JSON(res, 200, user)
	return
}

//Delete user
func (usersCtrl UsersCtrl) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	userID := vars["id"]

	if userID == "" {
		jsonErr := helpers.GenerateJSONError(helpers.ErrBadRequest, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	var user models.User
	if err := user.DeleteByID(userID); err != nil {
		jsonErr := helpers.GenerateJSONError(err, req.Header)
		r.JSON(res, jsonErr.StatusCode, jsonErr)
		return
	}

	r.Text(res, 204, "")
}
