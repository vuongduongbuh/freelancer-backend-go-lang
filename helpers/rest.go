package helpers

import (
	"net/http"

	njr "github.com/daseinhorn/negroni-json-recovery"
	"github.com/nicksnyder/go-i18n/i18n"
)

//TranslateStr wraps the i18n package for easier use
func TranslateStr(translationKey string, lang string) string {
	T, err := i18n.Tfunc(lang, "en-US")
	if err != nil {
		log.Error(err)
	}
	return T(translationKey)
}

//GenerateErrorResponse parses errors in a json format
func GenerateErrorResponse(translationKey string, reqHeader http.Header) njr.JSONPanicError {
	translatedError := TranslateStr(translationKey, reqHeader.Get("accept-language"))
	return njr.JSONPanicError{
		Code:   1000,
		Short:  translationKey,
		Errors: []interface{}{translatedError},
	}
}
