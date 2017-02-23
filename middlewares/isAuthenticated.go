package middlewares

import (
	"net/http"

	"errors"

	"github.com/auth0/go-jwt-middleware"
	"github.com/unrolled/render"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"golang.org/x/net/context"
)

//GetUserFromContext ensures a valid user is received from the context
func GetTokenFromContext(_ http.ResponseWriter, req *http.Request) (string, error) {
	token := req.Context().Value("token").(string)
	if token != "" {
		return token, nil
	}
	return "", errors.New("couldn't find token")
}

//IsAuthenticatedMiddleware passes user to the request context
func IsAuthenticatedMiddleware(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	req = req.WithContext(context.Background())

	token, err := jwtmiddleware.FromAuthHeader(req)
	if token != "" && err == nil {
		ctx := context.WithValue(req.Context(), "token", token)
		req = req.WithContext(ctx)
		next(res, req)
		return
	}

	r := render.New(render.Options{})
	r.JSON(res, 401, helpers.GenerateErrorResponse("invalid_token", req.Header))
	return
}
