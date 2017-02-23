package middlewares

import (
	"net/http"

	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/models"
	"golang.org/x/net/context"
)

//GetSurveyTicketFromContext ensures a valid ticket is received from the context
func GetSurveyTicketFromContext(req *http.Request) (models.SurveyTicket, error) {
	surveyTicket := req.Context().Value("surveyTicket").(models.SurveyTicket)
	if surveyTicket.ID.Valid() {
		return surveyTicket, nil
	}
	return models.SurveyTicket{}, models.ErrInvalidToken
}

//HasValidSurveyTicketMiddleware passes user to the request context
func HasValidSurveyTicketMiddleware(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	r := render.New(render.Options{})

	req = req.WithContext(context.Background())

	rawAuthHeader := req.Header.Get("Authorization")
	if rawAuthHeader == "" {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	var surveyTicket models.SurveyTicket
	if err := surveyTicket.FindByAuthHeader(rawAuthHeader); err != nil {
		r.JSON(res, 401, helpers.GenerateErrorResponse(models.ErrInvalidToken.Error(), req.Header))
		return
	}

	ctx := context.WithValue(req.Context(), "surveyTicket", surveyTicket)
	req = req.WithContext(ctx)
	next(res, req)
	return
}
