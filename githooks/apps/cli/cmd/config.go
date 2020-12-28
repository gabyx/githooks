package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manages various Githooks configuration.",
	Long: `
Manages various Githooks configuration.

git hooks config list [--local|--global]

	Lists the Githooks related settings of the Githooks configuration.
    Can be either global or local configuration, or both by default.

git hooks config [set|reset|print] disable

    Disables running any Githooks files in the current repository,
    when the 'set' option is used.
    The 'reset' option clears this setting.
    The 'print' option outputs the current setting.
    This command needs to be run at the root of a repository.

[deprecated] git hooks config [set|reset|print] single

    This command is deprecated and will be removed in the future.
    Marks the current local repository to be managed as a single Githooks
    installation, or clears the marker, with 'set' and 'reset' respectively.
    The 'print' option outputs the current setting of it.
    This command needs to be run at the root of a repository.

git hooks config set search-dir <path>
git hooks config [reset|print] search-dir

    Changes the previous search directory setting used during installation.
    The 'set' option changes the value, and the 'reset' option clears it.
    The 'print' option outputs the current setting of it.

git hooks config set shared [--local] <git-url...>
git hooks config [reset|print] shared [--local]

    Updates the list of global (or local) shared hook repositories when
    the 'set' option is used, which accepts multiple <git-url> arguments,
    each containing a clone URL of a hook repository.
    The 'reset' option clears this setting.
    The 'print' option outputs the current setting.

git hooks config [accept|deny|reset|print] trusted

    Accepts changes to all existing and new hooks in the current repository
    when the trust marker is present and the 'set' option is used.
    The 'deny' option marks the repository as
    it has refused to trust the changes, even if the trust marker is present.
    The 'reset' option clears this setting.
    The 'print' option outputs the current setting.
    This command needs to be run at the root of a repository.

git hooks config [enable|disable|reset|print] update

    Enables or disables automatic update checks with
    the 'enable' and 'disable' options respectively.
    The 'reset' option clears this setting.
    The 'print' option outputs the current setting.

git hooks config set clone-url <git-url>
git hooks config [set|print] clone-url

    Sets or prints the configured githooks clone url used
    for any update.

git hooks config set clone-branch <branch-name>
git hooks config print clone-branch

    Sets or prints the configured branch of the update clone
    used for any update.

git hooks config [reset|print] update-time

    Resets the last Githooks update time with the 'reset' option,
    causing the update check to run next time if it is enabled.
    Use 'git hooks update [enable|disable]' to change that setting.
    The 'print' option outputs the current value of it.

git hooks config [enable|disable|print] fail-on-non-existing-shared-hooks [--local|--global]

	Enable or disable failing hooks with an error when any
	shared hooks configured in '.shared' are missing,
	which usually means 'git hooks update' has not been called yet.

git hooks config [yes|no|reset|print] delete-detected-lfs-hooks

	By default, detected LFS hooks during install are disabled and backed up.
	The 'yes' option remembers to always delete these hooks.
	The 'no' option remembers the default behavior.
	The decision is reset with 'reset' to the default behavior.
	The 'print' option outputs the current behavior.`,
	Run: runConfig,
}

func runConfig(cmd *cobra.Command, args []string) {

}

func init() { // nolint: gochecknoinits
	rootCmd.AddCommand(configCmd)
}
