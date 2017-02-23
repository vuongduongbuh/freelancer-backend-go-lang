package controllers

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/tealeg/xlsx"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/models"
)

//ParticipantsCtrl is the controller for /companies
type ParticipantsCtrl struct{}

func (participantsCtrl ParticipantsCtrl) getLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"controller": "ParticipantsCtrl",
	})
}

//List all participants of a given company
func (participantsCtrl ParticipantsCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	companyID := vars["id"]
	participantID := vars["participantID"]

	var company models.Company
	if err := company.FindByID(companyID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	if participantID != "" {
		if isObjectID := bson.IsObjectIdHex(participantID); !isObjectID {
			r.JSON(res, 422, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
			return
		}

		participant, err := company.FindParticipantByObjectID(bson.ObjectIdHex(participantID))
		if err != nil {
			r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
			return
		}

		r.JSON(res, 200, participant)
		return
	}

	participants, err := company.FindParticipants()
	if err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	r.JSON(res, 200, participants)
	return
}

//Create a participant for a company
func (participantsCtrl ParticipantsCtrl) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	companyID := vars["id"]

	var company models.Company
	if err := company.FindByID(companyID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var participant models.Participant
	participant.Create()

	if err := decoder.Decode(&participant); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}
	createdParticipant, err := company.AddParticipant(participant)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 201, createdParticipant)
}

//CreateFromCSV allows to upload a CSV file which creates the participants in the background
func (participantsCtrl ParticipantsCtrl) CreateFromCSV(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	companyID := vars["id"]

	var company models.Company
	if err := company.FindByID(companyID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	req.ParseMultipartForm(0)
	participantsCSVFile, _, err := req.FormFile("file")
	if err != nil {
		return
	}
	defer participantsCSVFile.Close()

	//save file to disk to read again
	destinationFileName := "./assets/" + uuid.NewV4().String() + ".xlsx"

	destinationFile, err := os.OpenFile(destinationFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse("participants_malformed_csv", req.Header))
		return
	}
	defer destinationFile.Close()
	io.Copy(destinationFile, participantsCSVFile)

	defer func() {
		os.Remove(destinationFileName)
	}()

	//read file again to load excel
	xlFile, err := xlsx.OpenFile(destinationFileName)
	if err != nil {
		participantsCtrl.getLogger().WithFields(logrus.Fields{
			"function": "CreateFromCSV",
		}).Error(err)
		r.JSON(res, 422, helpers.GenerateErrorResponse("participants_malformed_csv", req.Header))
		return
	}

	if len(xlFile.Sheets) < 1 {
		r.JSON(res, 422, helpers.GenerateErrorResponse("participants_malformed_csv", req.Header))
		return
	}

	var participants []models.Participant

	firstSheet := xlFile.Sheets[0]
	for rowIdx, row := range firstSheet.Rows {
		if rowIdx == 0 {
			continue
		}

		var participant models.Participant
		participant.Create()

		for cellIdx, cell := range row.Cells {
			content, _ := cell.String()
			content = strings.Trim(content, " ")

			switch {
			case cellIdx == 0:
				participant.EMail = content
			case cellIdx == 1:
				participant.Language = content
			case cellIdx == 2:
				participant.FirstName = content
			case cellIdx == 3:
				participant.LastName = content
			}
		}

		participants = append(participants, participant)
	}

	if err := company.BulkAddParticipants(participants); err != nil {
		participantsCtrl.getLogger().Error(err)
	}

	r.Text(res, 204, "")
	return
}

//Delete a participant form a company
func (participantsCtrl ParticipantsCtrl) Delete(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	companyID := vars["id"]
	participantID := vars["participantID"]

	if isObjectID := bson.IsObjectIdHex(participantID); !isObjectID {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}

	var company models.Company
	if err := company.FindByID(companyID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	if err := company.DeleteParticipant(bson.ObjectIdHex(participantID)); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
}

//Update a participant for a company
func (participantsCtrl ParticipantsCtrl) Update(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	companyID := vars["id"]
	participantID := vars["participantID"]

	var company models.Company
	if err := company.FindByID(companyID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	if isObjectID := bson.IsObjectIdHex(participantID); !isObjectID {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrInvalidObjectID.Error(), req.Header))
		return
	}

	currentParticipant, err := company.FindParticipantByObjectID(bson.ObjectIdHex(participantID))
	if err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(models.ErrRecordNotFound.Error(), req.Header))
		return
	}

	updatedParticipant := currentParticipant

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&updatedParticipant); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	updatedParticipant.ID = currentParticipant.ID
	updatedParticipant.CreatedAt = currentParticipant.CreatedAt
	updatedParticipant.UpdatedAt = time.Now()

	if err := company.UpdateParticipant(updatedParticipant); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, updatedParticipant)
}
