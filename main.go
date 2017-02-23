package main

import (
	"encoding/base64"
	"errors"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/auth0/go-jwt-middleware"
	"github.com/codegangsta/negroni"
	"github.com/daseinhorn/negroni-json-recovery"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/rs/cors"
	"gitlab.com/slsurvey/slsurvey-srv/controllers"
	"gitlab.com/slsurvey/slsurvey-srv/helpers"
	"gitlab.com/slsurvey/slsurvey-srv/middlewares"
	"gitlab.com/slsurvey/slsurvey-srv/models"
	"gitlab.com/slsurvey/slsurvey-srv/validators"
)

var log *logrus.Logger

func isProdEnv() bool {
	return os.Getenv("APP_ENV") == "prod"
}

func bootstrap() {
	if isProdEnv() {
		err := godotenv.Load("prod.env", "main.env")
		if err != nil {
			panic(err.Error())
		}
	} else {
		err := godotenv.Load("main.env")
		if err != nil {
			panic(err.Error())
		}
	}

	log = helpers.GetLogger()

	validators.RegisterCustomValidators()

	// load translations
	i18n.MustLoadTranslationFile("assets/i18n/en-US.all.json")
}

func bootstrapCollection() {
	var participant models.Participant
	if err := participant.BootstrapCollection(); err != nil {
		panic(err)
	}
	var surveyTicket models.SurveyTicket
	if err := surveyTicket.BootstrapCollection(); err != nil {
		panic(err)
	}
}

func reportToSentry(err interface{}) {
	var panicErr error
	if e, ok := err.(error); ok {
		panicErr = e
	} else {
		panicErr = errors.New("negroni captured an error but was not able to convert interface to error")
	}
	log.Error(panicErr)
}

func main() {
	bootstrap()
	bootstrapCollection()

	decodedAuthSecret, decodingErr := base64.URLEncoding.DecodeString(os.Getenv("AUTH0_CLIENT_SECRET"))
	if decodingErr != nil {
		log.Panic(decodingErr)
	}

	//JWT middleware for auth0
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return decodedAuthSecret, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	baseRouter := mux.NewRouter()

	//controlers v1
	assetsCtrl := controllers.AssetsCtrl{}
	authCtrl := controllers.AuthCtrl{}
	companiesCtrl := controllers.CompaniesCtrl{}
	industriesCtrl := controllers.IndustriesCtrl{}
	usersCtrl := controllers.UsersCtrl{}
	participantsCtrl := controllers.ParticipantsCtrl{}
	surveyModuleCtrl := controllers.SurveyModulesCtrl{}
	surveyCatalogCtrl := controllers.SurveyCatalogCtrl{}
	surveyTicketsCtrl := controllers.TicketsCtrl{}
	conductCtrl := controllers.ConductCtrl{}
	userSettingsCtrl := controllers.UserSettingsCtrl{}
	explenationsCtrl := controllers.ExplenationsCtrl{}

	eMailTemplateCtrl := controllers.EmailTemplatesCtrl{}
	versionCtrl := controllers.VersionCtrl{}

	workbenchCtrl := controllers.WorkbenchCtrl{}

	supportCtrl := controllers.SupportCtrl{}

	//controllers v2
	surveyTicketsCtrlV2 := controllers.TicketsCtrlV2{}
	surveyConductCtrlV2 := controllers.ConductCtrlV2{}

	//auth API
	authRouter := mux.NewRouter()

	authRouter.HandleFunc("/api/v1/auth/check", authCtrl.Check).Methods("GET")

	baseRouter.PathPrefix("/api/v1/auth").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.HandlerFunc(middlewares.IsAuthenticatedMiddleware),
		negroni.Wrap(authRouter),
	))

	//assets API
	assetsRouter := mux.NewRouter()

	assetsRouter.HandleFunc("/api/v1/assets/images", assetsCtrl.UploadImages).Methods("POST")

	baseRouter.PathPrefix("/api/v1/assets").Handler(negroni.New(
		negroni.Wrap(assetsRouter),
	))

	//companies API
	companiesRouter := mux.NewRouter()

	companiesRouter.HandleFunc("/api/v1/companies", companiesCtrl.List).Methods("GET")
	companiesRouter.HandleFunc("/api/v1/companies", companiesCtrl.Create).Methods("POST")

	companiesRouter.HandleFunc("/api/v1/companies/{id}", companiesCtrl.List).Methods("GET")
	companiesRouter.HandleFunc("/api/v1/companies/{id}", companiesCtrl.Update).Methods("PUT")
	companiesRouter.HandleFunc("/api/v1/companies/{id}", companiesCtrl.Delete).Methods("DELETE")

	companiesRouter.HandleFunc("/api/v1/companies/{id}/participants", participantsCtrl.List).Methods("GET")
	companiesRouter.HandleFunc("/api/v1/companies/{id}/participants", participantsCtrl.Create).Methods("POST")
	companiesRouter.HandleFunc("/api/v1/companies/{id}/participants/csv", participantsCtrl.CreateFromCSV).Methods("POST")

	companiesRouter.HandleFunc("/api/v1/companies/{id}/participants/{participantID}", participantsCtrl.List).Methods("GET")
	companiesRouter.HandleFunc("/api/v1/companies/{id}/participants/{participantID}", participantsCtrl.Update).Methods("PUT")
	companiesRouter.HandleFunc("/api/v1/companies/{id}/participants/{participantID}", participantsCtrl.Delete).Methods("DELETE")

	baseRouter.PathPrefix("/api/v1/companies").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(companiesRouter),
	))

	//industries API
	industriesRouter := mux.NewRouter()

	industriesRouter.HandleFunc("/api/v1/industries", industriesCtrl.List).Methods("GET")

	baseRouter.PathPrefix("/api/v1/industries").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(industriesRouter),
	))

	//explenations API
	explenationsRouter := mux.NewRouter()

	explenationsRouter.HandleFunc("/api/v1/explenations", explenationsCtrl.List).Methods("GET")
	explenationsRouter.HandleFunc("/api/v1/explenations", explenationsCtrl.Create).Methods("POST")

	explenationsRouter.HandleFunc("/api/v1/explenations/{id}", explenationsCtrl.List).Methods("GET")
	explenationsRouter.HandleFunc("/api/v1/explenations/{id}", explenationsCtrl.Update).Methods("POST")
	explenationsRouter.HandleFunc("/api/v1/explenations/{id}", explenationsCtrl.Delete).Methods("DELETE")

	baseRouter.PathPrefix("/api/v1/explenations").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(explenationsRouter),
	))

	//surveyModuleRouter API
	surveyModuleRouter := mux.NewRouter()

	surveyModuleRouter.HandleFunc("/api/v1/modules", surveyModuleCtrl.List).Methods("GET")
	surveyModuleRouter.HandleFunc("/api/v1/modules", surveyModuleCtrl.Create).Methods("POST")

	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}", surveyModuleCtrl.List).Methods("GET")
	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}", surveyModuleCtrl.Update).Methods("PUT")
	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}", surveyModuleCtrl.Delete).Methods("DELETE")

	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}/publish", surveyModuleCtrl.Publish).Methods("PUT")
	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}/unpublish", surveyModuleCtrl.Unpublish).Methods("PUT")

	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}/questions", surveyModuleCtrl.AddQuestion).Methods("POST")
	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}/questions/sort", surveyModuleCtrl.SortQuestions).Methods("POST")
	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}/questions/{questionID}", surveyModuleCtrl.GetQuestion).Methods("GET")
	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}/questions/{questionID}", surveyModuleCtrl.UpdateQuestion).Methods("PUT")
	surveyModuleRouter.HandleFunc("/api/v1/modules/{id}/questions/{questionID}", surveyModuleCtrl.DeleteQuestion).Methods("DELETE")

	baseRouter.PathPrefix("/api/v1/modules").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(surveyModuleRouter),
	))

	//surveyCatalogRouter API
	surveyCatalogRouter := mux.NewRouter()

	surveyCatalogRouter.HandleFunc("/api/v1/catalogs", surveyCatalogCtrl.List).Methods("GET")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs", surveyCatalogCtrl.Create).Methods("POST")

	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}", surveyCatalogCtrl.List).Methods("GET")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}", surveyCatalogCtrl.Delete).Methods("DELETE")

	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/base", surveyCatalogCtrl.GetBaseSurveyModules).Methods("GET")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/base", surveyCatalogCtrl.AddBaseSurveyModules).Methods("POST")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/recurring", surveyCatalogCtrl.GetRecurringSurveyModules).Methods("GET")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/recurring", surveyCatalogCtrl.AddRecurringSurveyModules).Methods("POST")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/end", surveyCatalogCtrl.GetEndSurveyModules).Methods("GET")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/end", surveyCatalogCtrl.AddEndSurveyModules).Methods("POST")

	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/emailtemplate", surveyCatalogCtrl.SetEMailTemplateID).Methods("POST")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/settings", surveyCatalogCtrl.GetSettings).Methods("GET")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/settings", surveyCatalogCtrl.AddSettings).Methods("POST")
	surveyCatalogRouter.HandleFunc("/api/v1/catalogs/{catalogId}/publish", surveyCatalogCtrl.Publish).Methods("POST")

	baseRouter.PathPrefix("/api/v1/catalogs").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(surveyCatalogRouter),
	))

	//email template API
	emailRouter := mux.NewRouter()
	emailRouter.HandleFunc("/api/v1/emailtemplates", eMailTemplateCtrl.List).Methods("GET")
	emailRouter.HandleFunc("/api/v1/emailtemplates", eMailTemplateCtrl.Create).Methods("POST")

	emailRouter.HandleFunc("/api/v1/emailtemplates/{id}", eMailTemplateCtrl.List).Methods("GET")
	emailRouter.HandleFunc("/api/v1/emailtemplates/{id}", eMailTemplateCtrl.Update).Methods("PUT")
	emailRouter.HandleFunc("/api/v1/emailtemplates/{id}", eMailTemplateCtrl.Delete).Methods("DELETE")

	baseRouter.PathPrefix("/api/v1/emailtemplates").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(emailRouter),
	))

	//user API
	consultantsRouter := mux.NewRouter()
	consultantsRouter.HandleFunc("/api/v1/consultants", usersCtrl.List).Methods("GET")
	consultantsRouter.HandleFunc("/api/v1/consultants", usersCtrl.Create).Methods("POST")

	consultantsRouter.HandleFunc("/api/v1/consultants/{id}", usersCtrl.List).Methods("GET")
	consultantsRouter.HandleFunc("/api/v1/consultants/{id}", usersCtrl.Update).Methods("PATCH")
	consultantsRouter.HandleFunc("/api/v1/consultants/{id}", usersCtrl.Delete).Methods("DELETE")

	consultantsRouter.HandleFunc("/api/v1/consultants/{id}/toggleblock", usersCtrl.ToggleBlock).Methods("POST")

	baseRouter.PathPrefix("/api/v1/consultants").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(consultantsRouter),
	))

	// Mobile API (with tickets auth)

	//surveyTicket API
	surveyTicketRouter := mux.NewRouter()

	surveyTicketRouter.HandleFunc("/api/v1/ticket", surveyTicketsCtrl.List).Methods("GET")

	baseRouter.PathPrefix("/api/v1/ticket").Handler(negroni.New(
		negroni.HandlerFunc(middlewares.HasValidSurveyTicketMiddleware),
		negroni.Wrap(surveyTicketRouter),
	))

	//support API
	supportRouter := mux.NewRouter()
	supportRouter.HandleFunc("/api/v1/support/{id}/tickets", supportCtrl.LightTicketsList).Methods("GET")

	baseRouter.PathPrefix("/api/v1/support").Handler(negroni.New(
		negroni.HandlerFunc(middlewares.HasValidSurveyTicketMiddleware),
		negroni.Wrap(supportRouter),
	))

	//surveyTicket V2 API
	surveyTicketV2Router := mux.NewRouter()

	surveyTicketV2Router.HandleFunc("/api/v2/ticket", surveyTicketsCtrlV2.List).Methods("GET")

	baseRouter.PathPrefix("/api/v2/ticket").Handler(negroni.New(
		negroni.HandlerFunc(middlewares.HasValidSurveyTicketMiddleware),
		negroni.Wrap(surveyTicketV2Router),
	))

	//API to conduct a given survey
	surveyConductRouter := mux.NewRouter()

	surveyConductRouter.HandleFunc("/api/v1/conduct/{id}", conductCtrl.GetAnswers).Methods("GET")
	surveyConductRouter.HandleFunc("/api/v1/conduct/{id}", conductCtrl.Conduct).Methods("POST")
	surveyConductRouter.HandleFunc("/api/v1/conduct/{id}/submit", conductCtrl.Submit).Methods("POST")

	baseRouter.PathPrefix("/api/v1/conduct").Handler(negroni.New(
		negroni.HandlerFunc(middlewares.HasValidSurveyTicketMiddleware),
		negroni.Wrap(surveyConductRouter),
	))

	//surveyConduct v2 API
	surveyConductV2Router := mux.NewRouter()

	surveyConductV2Router.HandleFunc("/api/v2/conduct/{id}", surveyConductCtrlV2.GetAnswers).Methods("GET")
	surveyConductV2Router.HandleFunc("/api/v2/conduct/{id}", surveyConductCtrlV2.Conduct).Methods("POST")
	surveyConductV2Router.HandleFunc("/api/v2/conduct/{id}/submit", surveyConductCtrlV2.Submit).Methods("POST")

	baseRouter.PathPrefix("/api/v2/conduct").Handler(negroni.New(
		negroni.HandlerFunc(middlewares.HasValidSurveyTicketMiddleware),
		negroni.Wrap(surveyConductV2Router),
	))

	//user settings (language, pushtoken) API
	userSettingsRouter := mux.NewRouter()

	userSettingsRouter.HandleFunc("/api/v1/usersettings", userSettingsCtrl.Save).Methods("POST")
	userSettingsRouter.HandleFunc("/api/v1/usersettings", userSettingsCtrl.Delete).Methods("DELETE")

	baseRouter.PathPrefix("/api/v1/usersettings").Handler(negroni.New(
		negroni.HandlerFunc(middlewares.HasValidSurveyTicketMiddleware),
		negroni.Wrap(userSettingsRouter),
	))

	//version API
	versionRouter := mux.NewRouter()

	versionRouter.HandleFunc("/api/v1/version", versionCtrl.List).Methods("GET")

	baseRouter.PathPrefix("/api/v1/version").Handler(negroni.New(
		negroni.Wrap(versionRouter),
	))

	//workbench API
	workbenchRouter := mux.NewRouter()

	workbenchRouter.HandleFunc("/api/v1/workbench", workbenchCtrl.Test).Methods("GET")

	baseRouter.PathPrefix("/api/v1/workbench").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(workbenchRouter),
	))

	//negroni routing setup
	n := negroni.New()

	sentryRecovery := negroni.NewRecovery()
	sentryRecovery.ErrorHandlerFunc = reportToSentry
	n.Use(sentryRecovery)

	//disable logging if not in dev mode
	if !isProdEnv() || true {
		n.Use(recovery.JSONRecovery(true))
		n.Use(negroni.NewLogger())
	}

	// configure cors settings
	options := cors.Options{
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
	}

	if isProdEnv() {
		options.AllowedOrigins = []string{
			//prod configs
			"https://nowatwork.ch",
			"https://control.nowatwork.ch",
			"https://naw-control.ch",
		}
		options.Debug = false
	} else {
		options.AllowedOrigins = []string{"*"}
		options.Debug = false
	}
	n.Use(cors.New(options))

	n.UseHandler(baseRouter)

	// Get PORT from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "3008"
	}

	n.Run(":" + port)
}
