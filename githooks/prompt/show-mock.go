// +build mock

package prompt

import (
	"os"
	"strings"
)

const AssertOutputIsTerminal = false

var EnableGUI bool = os.Getenv("GH_ENABLE_GUI") == "true"

// ShowOptions mocks the real ShowOptions by reading
// from the environment or if not defined calls the normal implementation.
// This is only for tests.
func (p *Context) ShowOptions(text string,
	hintText string,
	shortOptions string,
	longOptions ...string) (answer string, err error) {

	if strings.Contains(text, "This repository wants you to trust all current") {
		answer, defined := os.LookupEnv("TRUST_ALL_HOOKS")
		if defined {
			return strings.ToLower(answer), nil
		}
	} else if strings.Contains(text, "Do you accept the changes") {
		answer, defined := os.LookupEnv("ACCEPT_CHANGES")
		if defined {
			return strings.ToLower(answer), nil
		}
	} else if strings.Contains(text, "There is a new Githooks update available") {
		answer, defined := os.LookupEnv("EXECUTE_UPDATE")
		if defined {
			return strings.ToLower(answer), nil
		}
	}

	return showOptions(p, text, hintText, shortOptions, longOptions...)
}

// ShowEntry mocks the real ShowPrompt by reading
// from the environment or if not defined calls the normal implementation.
// This is only for tests.
func (p *Context) ShowEntry(
	text string,
	defaultAnswer string,
	validator AnswerValidator) (answer string, err error) {
	return showEntry(p, text, defaultAnswer, validator, false)
}

// ShowEntryMulti shows multiple prompts to enter multiple answers and
// validates it with a validator. An empty answer exits the prompt.
func (p *Context) ShowEntryMulti(
	text string,
	exitAnswer string,
	validator AnswerValidator) (answers []string, err error) {
	return showEntryMulti(p, text, exitAnswer, validator)
}
