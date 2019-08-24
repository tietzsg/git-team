package disablecmdadapter

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/hekmekk/git-team/src/core/config"
	"github.com/hekmekk/git-team/src/core/gitconfig"
	"github.com/hekmekk/git-team/src/core/state_repository"
	"github.com/hekmekk/git-team/src/disable"
)

// Definition the command, arguments, and dependencies
type Definition struct {
	CommandName string
	Policy      disable.Policy
}

// NewDefinition the constructor for Definition
func NewDefinition(app *kingpin.Application) Definition {

	command := app.Command("disable", "Use default commit template and remove prepare-commit-msg hook")

	return Definition{
		CommandName: command.FullCommand(),
		Policy: disable.Policy{
			Deps: disable.Dependencies{
				GitUnsetCommitTemplate: gitconfig.UnsetCommitTemplate,
				GitUnsetHooksPath:      gitconfig.UnsetHooksPath,
				LoadConfig:             config.Load,
				StatFile:               os.Stat,
				RemoveFile:             os.Remove,
				PersistDisabled:        staterepository.PersistDisabled,
			},
		},
	}
}
