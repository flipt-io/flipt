package util

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// PromptPlaintext prompts a user to input plain text
func PromptPlaintext(message string, defaultVal string) (string, error) {
	var (
		prompt = survey.Input{
			Message: message,
			Default: defaultVal,
		}

		value string
	)

	if err := survey.AskOne(&prompt, &value, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return strings.TrimSpace(value), nil

}

// PromptPassword prompts a user to input a hidden field
func PromptPassword(message string) (string, error) {
	var (
		prompt = &survey.Password{Message: message}

		password string
	)

	if err := survey.AskOne(prompt, &password, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return strings.TrimSpace(password), nil
}

func PromptConfirm(message string, defaultVal bool) (bool, error) {
	var (
		prompt = &survey.Confirm{
			Message: message,
			Default: defaultVal,
		}

		value bool
	)

	if err := survey.AskOne(prompt, &value); err != nil {
		return false, err
	}

	return value, nil
}
