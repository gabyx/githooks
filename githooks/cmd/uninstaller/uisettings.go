package uninstaller

import (
	"github.com/gabyx/githooks/githooks/prompt"
)

// UISettings defines user interface settings made by the user over prompts.
type UISettings struct {

	// A prompt context which enables showing a prompt.
	PromptCtx prompt.IContext
}
