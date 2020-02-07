package config

import (
	"fmt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	ens "github.com/go-playground/validator/translations/en"
	"gopkg.in/go-playground/validator.v9"
)

// Validate
func validate(config Config) error {
	validate := validator.New()

	// Get English Errors
	uni := ut.New(en.New(), en.New())
	trans, _ := uni.GetTranslator("en")
	_ = ens.RegisterDefaultTranslations(validate, trans)

	// Validate
	err := validate.Struct(&config)

	if err != nil {
		errorLists, ok := err.(validator.ValidationErrors)
		if ok {
			return fmt.Errorf("errors validating struct config: %v", errorLists.Translate(trans))
		}

		return fmt.Errorf("errors validating struct config: %v", err.Error())
	}

	errList := customValidate(config)
	if len(errList) != 0 {
		return fmt.Errorf("errors validating struct config: %v", errList)
	}

	return nil
}

// Custom Validators
func customValidate(config Config) []error {
	errList := make([]error, 0)

	// Put Custom Validation Here

	return errList
}
