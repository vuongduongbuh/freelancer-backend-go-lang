package helpers

import (
	"errors"
	"net/http"

	"github.com/nicksnyder/go-i18n/i18n"
)

var (
	//ErrBadRequest is used for 400
	ErrBadRequest = errors.New("error.bad-request")
	//ErrInvalidObjectID is used for 400
	ErrInvalidObjectID = errors.New("error.bad-request")
	//ErrUnauthorized is used for 401
	ErrUnauthorized = errors.New("error.unauthorized")
	//ErrForbidden is used for 403
	ErrForbidden = errors.New("error.forbidden")
	//ErrRecordNotFound is used for 404
	ErrRecordNotFound = errors.New("error.record-not-found")
	//ErrRequestTimeOut is used for 408
	ErrRequestTimeOut = errors.New("error.request-time-out")
	//ErrUnprocessableEntity is used for 422
	ErrUnprocessableEntity = errors.New("error.unprocessable-entity")
	//ErrInternalServerError is used for 500
	ErrInternalServerError = errors.New("error.internal-server-error")

	//Auth0 errors

	//ErrAuth0IDP duplicated user
	ErrAuth0IDP = errors.New("auth0_idp_error")
)

//JSONError is used as error representation vor all v2 APIs
type JSONError struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
	ErrorCode  string `json:"errorCode"`
}

//translation function used to translate error and message

type jsonErrorTranslation struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func translateError(translationKey string, lang string) (msgs jsonErrorTranslation) {
	T, err := i18n.Tfunc(lang, "en-US")
	if err != nil {
		return jsonErrorTranslation{}
	}

	return jsonErrorTranslation{
		Error:   T(translationKey),
		Message: T(translationKey + "_msg"),
	}
}

func determineErrorCode(err error) int {
	code := 500

	switch err.Error() {
	case ErrBadRequest.Error():
		code = http.StatusBadRequest
	case ErrInvalidObjectID.Error():
		code = http.StatusBadRequest
	case ErrAuth0IDP.Error():
		code = http.StatusBadRequest
	case ErrRecordNotFound.Error():
		code = http.StatusNotFound
	case ErrUnauthorized.Error():
		code = http.StatusUnauthorized
	case ErrUnprocessableEntity.Error():
		code = http.StatusUnprocessableEntity
	}

	return code
}

//GenerateJSONError parses errors and returns a rest error
func GenerateJSONError(err error, reqHeader http.Header) JSONError {
	errorCode := err.Error()
	msgs := translateError(errorCode, reqHeader.Get("accept-language"))

	return JSONError{
		StatusCode: determineErrorCode(err),
		Error:      msgs.Error,
		Message:    msgs.Message,
		ErrorCode:  errorCode,
	}
}
