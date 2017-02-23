package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/mgo.v2/bson"
)

const (
	//UserModelRoleParticipant for general low access roles
	UserModelRoleParticipant = 1
	//UserModelRoleConsultant for general genearl consultant access
	UserModelRoleConsultant = 2
)

var (
	UserGroupNameAdmin      = "admin"
	UserGroupNameSuperAdmin = "headbits"
)

//User model
type User struct {
	ID string `json:"id"`

	Picture string `json:"picture"`

	CompanyObjectID bson.ObjectId `json:"companyId"`
	CompanyName     string        `json:"companyName"`

	EMail         string `json:"email"`
	EMailVerified bool   `json:"emailVerified"`

	LastIP      string    `json:"lastIP"`
	LastLogin   time.Time `json:"lastLogin"`
	LoginsCount int       `json:"loginsCount"`

	IsAdmin      bool `json:"isAdmin"`
	IsSuperAdmin bool `json:"isSuperAdmin"`
	Blocked      bool `json:"blocked"`

	UserMetadata userAuth0UserMetadata `json:"-"`
}

//convertFromAuth0 users
func (user *User) convertFromAuth0(auth0User userAuth0UserResponseModel) error {
	user.ID = auth0User.UserID

	user.CompanyObjectID = auth0User.UserMetadata.CompanyObjectID
	user.CompanyName = auth0User.UserMetadata.CompanyName

	user.Picture = auth0User.Picture

	user.EMail = auth0User.Email
	user.EMailVerified = auth0User.EmailVerified

	user.LastIP = auth0User.LastIP
	user.LastLogin = auth0User.LastLogin
	user.LoginsCount = auth0User.LoginsCount

	user.Blocked = auth0User.Blocked

	for _, group := range auth0User.AppMetadata.Authorization.Groups {
		if group == UserGroupNameAdmin {
			user.IsAdmin = true
		}
		if group == UserGroupNameSuperAdmin {
			user.IsSuperAdmin = true
		}
	}

	user.UserMetadata = auth0User.UserMetadata

	return nil
}

//getLogger
func (user User) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "User",
	})
}

var auth0ManagementToken = ""

type userAuth0TokenResponseModel struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

//getAuth0Token returns JWT token for admin access
func getAuth0Token(forceRefresh bool) (string, error) {
	if auth0ManagementToken != "" && !forceRefresh {
		return auth0ManagementToken, nil
	}

	url := "https://nowatwork.eu.auth0.com/oauth/token"
	payload := strings.NewReader("{\"client_id\":\"3q9JMsbrFjOk43JLgF4jgh5KapusR42c\",\"client_secret\":\"uRD7l5NwoCeFxHO8yno_bstLJZGo-4dqRkT3lkPOPRTXK_dov8s3RugLuN2VSVYF\",\"audience\":\"https://nowatwork.eu.auth0.com/api/v2/\",\"grant_type\":\"client_credentials\"}")
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return "", err
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var token userAuth0TokenResponseModel

	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&token); err != nil {
		return "", err
	}

	parsedToken := token.TokenType + " " + token.AccessToken
	return parsedToken, nil
}

//checks response for error
func checkAuth0Response(url string, res *http.Response) error {
	statusCode := res.StatusCode

	if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusNoContent {
		return nil
	}

	if statusCode == http.StatusNotFound {
		return helpers.ErrRecordNotFound
	}
	if statusCode == http.StatusUnauthorized {
		getLogger().WithFields(logrus.Fields{
			"URL": url,
		}).Error(helpers.ErrUnauthorized)
		return helpers.ErrUnauthorized
	}
	if statusCode == http.StatusForbidden {
		return helpers.ErrForbidden
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	getLogger().WithFields(logrus.Fields{
		"URL": url,
	}).Error(string(body))
	return helpers.ErrRequestTimeOut
}

//Validate the given question
func (user *User) Validate() error {
	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		return err
	}

	return nil
}

type userAuth0UserMetadata struct {
	CompanyObjectID bson.ObjectId `json:"companyObjectId,omitempty"`
	CompanyName     string        `json:"companyName,omitempty"`
}

type userAuth0UserResponseModel struct {
	EmailVerified bool      `json:"email_verified"`
	Email         string    `json:"email"`
	UpdatedAt     time.Time `json:"updated_at"`
	Name          string    `json:"name"`
	Picture       string    `json:"picture"`
	UserID        string    `json:"user_id"`
	Nickname      string    `json:"nickname"`
	Identities    []struct {
		UserID     string `json:"user_id"`
		Provider   string `json:"provider"`
		Connection string `json:"connection"`
		IsSocial   bool   `json:"isSocial"`
	} `json:"identities"`
	CreatedAt   time.Time `json:"created_at"`
	LastIP      string    `json:"last_ip"`
	LastLogin   time.Time `json:"last_login"`
	LoginsCount int       `json:"logins_count"`
	Blocked     bool      `json:"blocked"`
	AppMetadata struct {
		Authorization struct {
			Groups []string `json:"groups"`
		} `json:"authorization"`
	} `json:"app_metadata"`
	UserMetadata userAuth0UserMetadata `json:"user_metadata"`
}

//FindAll users
func (user *User) FindAll() ([]User, error) {
	url := "https://nowatwork.eu.auth0.com/api/v2/users"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	authToken, err := getAuth0Token(false)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", authToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if err := checkAuth0Response(url, res); err != nil {
		return nil, err
	}

	var auth0Users []userAuth0UserResponseModel

	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&auth0Users); err != nil {
		return nil, err
	}

	var users []User
	for _, auth0User := range auth0Users {
		var user User
		user.convertFromAuth0(auth0User)
		users = append(users, user)
	}

	return users, nil
}

//FindByID user
func (user *User) FindByID(id string) error {
	url := "https://nowatwork.eu.auth0.com/api/v2/users/" + url.QueryEscape(id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	authToken, err := getAuth0Token(false)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", authToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if err := checkAuth0Response(url, res); err != nil {
		return err
	}

	var auth0User userAuth0UserResponseModel

	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&auth0User); err != nil {
		return err
	}

	user.convertFromAuth0(auth0User)

	return nil
}

type userAuth0UserCreateRequestModel struct {
	Connection    string                `json:"connection"`
	EMail         string                `json:"email"`
	EMailVerified bool                  `json:"email_verified"`
	Password      string                `json:"password"`
	UserMetadata  userAuth0UserMetadata `json:"user_metadata,omitempty"`
}

func generateRandomPassword() string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 12)
	for i := 0; i < 8; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

//Create a user at auth0
func (user *User) Create(eMail string, password string, companyObjectID bson.ObjectId) error {
	url := "https://nowatwork.eu.auth0.com/api/v2/users"

	payloadData := userAuth0UserCreateRequestModel{
		Connection:    "Username-Password-Authentication",
		EMail:         eMail,
		EMailVerified: true,
	}

	//check for companyId
	if !companyObjectID.Valid() {
		return helpers.ErrBadRequest
	}

	var company Company
	if err := company.FindByObjectID(companyObjectID); err != nil {
		return helpers.ErrRecordNotFound
	}
	payloadData.UserMetadata.CompanyObjectID = companyObjectID
	payloadData.UserMetadata.CompanyName = company.Name

	//generate password if empty
	if password == "" {
		password = generateRandomPassword()
	}
	payloadData.Password = password

	payload, err := json.Marshal(payloadData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	authToken, err := getAuth0Token(false)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", authToken)
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusCreated {
		var auth0User userAuth0UserResponseModel

		if err := json.NewDecoder(res.Body).Decode(&auth0User); err != nil {
			return err
		}

		user.convertFromAuth0(auth0User)
		return nil
	}

	var auth0Err helpers.JSONError
	if err := json.NewDecoder(res.Body).Decode(&auth0Err); err != nil {
		return err
	}

	user.getLogger().WithFields(logrus.Fields{
		"URL": url,
	}).Error(auth0Err)

	return errors.New(auth0Err.ErrorCode)
}

type userAuth0UserUpdateRequestModel struct {
	Blocked           bool                  `json:"blocked"`
	EmailVerified     bool                  `json:"email_verified,omitempty"`
	Email             string                `json:"email,omitempty"`
	VerifyEmail       bool                  `json:"verify_email,omitempty"`
	PhoneNumber       string                `json:"phone_number,omitempty"`
	PhoneVerified     bool                  `json:"phone_verified,omitempty"`
	VerifyPhoneNumber bool                  `json:"verify_phone_number,omitempty"`
	Password          string                `json:"password,omitempty"`
	VerifyPassword    bool                  `json:"verify_password,omitempty"`
	UserMetadata      userAuth0UserMetadata `json:"user_metadata,omitempty"`
	Connection        string                `json:"connection,omitempty"`
	Username          string                `json:"username,omitempty"`
	ClientID          string                `json:"client_id,omitempty"`
}

//UserUpdateRequestModel is used as abstraction between the client facting API and auth0
type UserUpdateRequestModel struct {
	CompanyObjectID bson.ObjectId `json:"companyId,omitempty"`
	Blocked         bool          `json:"blocked,omitempty"`
	Email           string        `json:"email,omitempty"`
	Password        string        `json:"password,omitempty"`
}

//ToogleBlock user
func (user *User) ToogleBlock() error {
	if user.IsAdmin || user.IsSuperAdmin {
		return errors.New("cant-deactivate-admin")
	}

	reqData := userAuth0UserUpdateRequestModel{}
	reqData.Blocked = !user.Blocked
	reqData.UserMetadata = user.UserMetadata

	url := "https://nowatwork.eu.auth0.com/api/v2/users/" + url.QueryEscape(user.ID)

	payload, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	authToken, err := getAuth0Token(false)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", authToken)
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err := checkAuth0Response(url, res); err != nil {
		return err
	}

	var auth0User userAuth0UserResponseModel
	if err := json.NewDecoder(res.Body).Decode(&auth0User); err != nil {
		return err
	}

	user.convertFromAuth0(auth0User)
	return nil
}

//Update user
func (user *User) Update(updateRequest UserUpdateRequestModel) error {
	reqData := userAuth0UserUpdateRequestModel{}

	reqData.Blocked = user.Blocked

	//check for companyId
	if updateRequest.CompanyObjectID.Valid() {
		var company Company
		if err := company.FindByObjectID(updateRequest.CompanyObjectID); err != nil {
			return helpers.ErrRecordNotFound
		}
		reqData.UserMetadata.CompanyObjectID = updateRequest.CompanyObjectID
		reqData.UserMetadata.CompanyName = company.Name
	}

	//check if EMail has changed
	if user.EMail != updateRequest.Email {
		reqData.Email = updateRequest.Email
	}

	if updateRequest.Password != "" {
		reqData.Password = updateRequest.Password
	}

	url := "https://nowatwork.eu.auth0.com/api/v2/users/" + url.QueryEscape(user.ID)

	payload, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	authToken, err := getAuth0Token(false)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", authToken)
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err := checkAuth0Response(url, res); err != nil {
		return err
	}

	var auth0User userAuth0UserResponseModel
	if err := json.NewDecoder(res.Body).Decode(&auth0User); err != nil {
		return err
	}

	user.convertFromAuth0(auth0User)
	return nil
}

//Delete a user
func (user *User) Delete() error {
	return user.DeleteByID(user.ID)
}

//DeleteByID a user by id
func (user User) DeleteByID(userID string) error {
	url := "https://nowatwork.eu.auth0.com/api/v2/users/" + url.QueryEscape(userID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	authToken, err := getAuth0Token(false)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", authToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if err := checkAuth0Response(url, res); err != nil {
		return err
	}

	return nil
}
