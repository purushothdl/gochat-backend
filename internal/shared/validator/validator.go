// internal/shared/validator/validator.go
package validator

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	v.RegisterValidation("password", validatePassword)
	return &Validator{validate: v}
}

func (v *Validator) Validate(data any) map[string]string {
	err := v.validate.Struct(data)
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors)
	errorMessages := make(map[string]string)

	for _, fieldErr := range validationErrors {
		fieldName := fieldErr.Field()
		switch fieldErr.Tag() {
		case "required":
			errorMessages[fieldName] = fmt.Sprintf("The %s field is required.", fieldName)
		case "email":
			errorMessages[fieldName] = fmt.Sprintf("The %s field must be a valid email address.", fieldName)
		case "min":
			errorMessages[fieldName] = fmt.Sprintf("The %s field must be at least %s characters long.", fieldName, fieldErr.Param())
		case "url":
			errorMessages[fieldName] = fmt.Sprintf("The %s field must be a valid URL.", fieldName)
		case "oneof":
			allowedValues := strings.ReplaceAll(fieldErr.Param(), " ", " | ")
			errorMessages[fieldName] = fmt.Sprintf("The %s field must be one of the following: %s.", fieldName, allowedValues)
		case "password":
			errorMessages[fieldName] = "The password must contain at least one uppercase letter, one lowercase letter, one number, and one special character."
		default:
			errorMessages[fieldName] = fmt.Sprintf("The %s field is not valid.", fieldName)
		}
	}

	return errorMessages
}

func validatePassword(fl validator.FieldLevel) bool {
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	password := fl.Field().String()

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}
