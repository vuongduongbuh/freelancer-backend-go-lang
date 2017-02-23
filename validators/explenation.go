package validators

import (
	"errors"

	govalidator "gopkg.in/asaskevich/govalidator.v4"

	"gitlab.com/slsurvey/slsurvey-srv/models"
)

func explenationCast(v interface{}) (models.Explenation, error) {
	explenation, isCorrectType := v.(models.Explenation)
	if !isCorrectType {
		return models.Explenation{}, errors.New("wrong type")
	}
	return explenation, nil
}

func registerExplenationValidator() {
	govalidator.CustomTypeTagMap.Set("explenationContent", func(i interface{}, o interface{}) bool {
		if explenation, err := explenationCast(o); err == nil {
			return validateExplenation(explenation)
		}
		return false
	})
}

func validateExplenation(explenation models.Explenation) bool {
	languages := make(map[string]bool)
	for _, explenationContent := range explenation.Content {
		if languages[explenationContent.Language] {
			return false
		}
		languages[explenationContent.Language] = true
	}
	return true
}
