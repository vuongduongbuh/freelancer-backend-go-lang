package validators

//RegisterCustomValidators is the init function
//for all the custom validators
func RegisterCustomValidators() {
	registerQuestionValidators()
	registerSurveyCatalogValidators()
	registerExplenationValidator()
}
