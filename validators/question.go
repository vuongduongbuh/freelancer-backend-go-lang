package validators

import (
	"errors"

	"gitlab.com/slsurvey/slsurvey-srv/models"
	"gopkg.in/asaskevich/govalidator.v4"
)

func registerQuestionValidators() {
	registerQuestionKeyboardValidator()
	registerQuestionSPSSIDValidator()
}

func questionCast(v interface{}) (models.Question, error) {
	question, isCorrectType := v.(models.Question)
	if !isCorrectType {
		return models.Question{}, errors.New("wrong type")
	}
	return question, nil
}

func registerQuestionSPSSIDValidator() {
	govalidator.CustomTypeTagMap.Set("spssid", func(i interface{}, o interface{}) bool {
		if question, err := questionCast(o); err == nil {
			return question.IsSPSSIDUnique()
		}
		return false
	})
}

func registerQuestionKeyboardValidator() {
	govalidator.CustomTypeTagMap.Set("questionKeyboard", func(i interface{}, o interface{}) bool {
		if question, err := questionCast(o); err == nil {
			if question.Keyboard.Type == models.QuestionKeyboardTypeDefault {
				return validateQuestionTypeDefault(question)
			}
			if question.Keyboard.Type == models.QuestionKeyboardTypeText {
				return validateQuestionTypeText(question)
			}
			if question.Keyboard.Type == models.QuestionKeyboardTypeInformation {
				return validateQuestionTypeInformation(question)
			}
			if question.Keyboard.Type == models.QuestionKeyboardTypeSmileboard {
				return true
			}
			return false
		}
		return false
	})
}

//Question validators

// Validate question typoe default
func validateQuestionTypeDefault(question models.Question) bool {
	keyboard := question.Keyboard

	// keys are required
	if len(keyboard.Keys) < 1 {
		return false
	}

	//validate keyboard keys
	requiredLanguages := make(map[string]bool, 0)
	for _, key := range keyboard.Keys {
		_, err := govalidator.ValidateStruct(key)
		if err != nil {
			return false
		}
		for _, localizedText := range key.Text {
			requiredLanguages[localizedText.Language] = true
		}
	}

	for _, key := range keyboard.Keys {
		implementedLanguages := make(map[string]bool, 0)

		for _, localizedText := range key.Text {
			implementedLanguages[localizedText.Language] = true
		}
		if len(implementedLanguages) != len(requiredLanguages) {
			return false
		}
	}

	return true
}

// Validate question type text
func validateQuestionTypeText(question models.Question) bool {
	keyboard := question.Keyboard

	// keys are not allowed here
	if len(keyboard.Keys) > 0 {
		return false
	}

	// placeholders are required
	if len(keyboard.Placeholder) < 1 {
		return false
	}

	// validate min and max
	if keyboard.Max > 0 {
		if keyboard.Min > keyboard.Max {
			return false
		}
	}

	// validate langauges
	requiredLanguages := make(map[string]bool, 0)
	for _, placeholder := range question.Keyboard.Placeholder {
		requiredLanguages[placeholder.Language] = true
	}

	for _, text := range question.Text {
		if !requiredLanguages[text.Language] {
			return false
		}
	}

	return true
}

// Validate question type information
func validateQuestionTypeInformation(question models.Question) bool {
	keyboard := question.Keyboard

	// keys are not allowed here
	if len(keyboard.Keys) > 0 {
		return false
	}

	// title is required
	if len(keyboard.Title) < 1 {
		return false
	}
	// button is required
	if len(keyboard.Button) < 1 {
		return false
	}

	// validate langauges
	requiredLanguages := make(map[string]bool, 0)
	for _, title := range question.Keyboard.Title {
		requiredLanguages[title.Language] = true
	}
	for _, button := range question.Keyboard.Button {
		if !requiredLanguages[button.Language] {
			return false
		}
	}
	for _, text := range question.Text {
		if !requiredLanguages[text.Language] {
			return false
		}
	}

	return true
}
