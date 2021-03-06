package hooks

// Git config keys for globals config.
const (
	GitCKInstallDir = "githooks.installDir"
	GitCKRunner     = "githooks.runner"
	GitCKDialog     = "githooks.dialog"

	GitCKDisable = "githooks.disable"

	GitCKMaintainOnlyServerHooks = "githooks.maintainOnlyServerHooks"

	GitCKAutoUpdateEnabled        = "githooks.autoUpdateEnabled"
	GitCKAutoUpdateCheckTimestamp = "githooks.autoUpdateCheckTimestamp"
	GitCKAutoUpdateUsePrerelease  = "githooks.autoUpdateUsePrerelease"

	GitCKBugReportInfo = "githooks.bugReportInfo"

	GitCKChecksumCacheDir = "githooks.checksumCacheDir"

	GitCKCloneBranch     = "githooks.cloneBranch"
	GitCKCloneURL        = "githooks.cloneUrl"
	GitCKBuildFromSource = "githooks.buildFromSource"
	GitCKGoExecutable    = "githooks.goExecutable"

	GitCKDeleteDetectedLFSHooksAnswer = "githooks.deleteDetectedLFSHooks"

	GitCKUseCoreHooksPath        = "githooks.useCoreHooksPath"
	GitCKPathForUseCoreHooksPath = "githooks.pathForUseCoreHooksPath"

	GitCKPreviousSearchDir = "githooks.previousSearchDir"
	GitCKNumThreads        = "githooks.numThreads"

	GitCKAliasHooks = "alias.hooks"
)

// Git config keys for local config.
const (
	GitCKRegistered = "githooks.registered"
	GitCKTrustAll   = "githooks.trustAll"
)

// Git config keys for local/global config.
const (
	GitCKShared                        = "githooks.shared"
	GitCKSharedUpdateTriggers          = "githooks.sharedHooksUpdateTriggers"
	GitCKAutoUpdateSharedHooksDisabled = "githooks.autoUpdateSharedHooksDisabled"

	GitCKSkipNonExistingSharedHooks = "githooks.skipNonExistingSharedHooks"
	GitCKSkipUntrustedHooks         = "githooks.skipUntrustedHooks"

	GitCKRunnerIsNonInteractive = "githooks.runnerIsNonInteractive"
)

// GetGlobalGitConfigKeys gets all global git config keys relevant for Githooks.
func GetGlobalGitConfigKeys() []string {
	return []string{
		GitCKInstallDir,
		GitCKRunner,
		GitCKDialog,

		GitCKDisable,

		GitCKMaintainOnlyServerHooks,
		GitCKPreviousSearchDir,

		GitCKAutoUpdateEnabled,
		GitCKAutoUpdateCheckTimestamp,
		GitCKAutoUpdateUsePrerelease,

		GitCKBugReportInfo,

		GitCKChecksumCacheDir,

		GitCKCloneBranch,
		GitCKCloneURL,
		GitCKGoExecutable,
		GitCKBuildFromSource,

		GitCKDeleteDetectedLFSHooksAnswer,

		GitCKUseCoreHooksPath,
		GitCKPathForUseCoreHooksPath,

		GitCKNumThreads,

		GitCKAliasHooks,

		// Local & global.
		GitCKShared,
		GitCKSharedUpdateTriggers,
		GitCKAutoUpdateSharedHooksDisabled,

		GitCKSkipNonExistingSharedHooks,
		GitCKSkipUntrustedHooks,

		GitCKRunnerIsNonInteractive}
}

// GetLocalGitConfigKeys gets all local git config keys relevant for Githooks.
func GetLocalGitConfigKeys() []string {
	return []string{
		GitCKRegistered,
		GitCKTrustAll,

		GitCKShared,
		GitCKSharedUpdateTriggers,
		GitCKAutoUpdateSharedHooksDisabled,

		GitCKSkipNonExistingSharedHooks,
		GitCKSkipUntrustedHooks,

		GitCKRunnerIsNonInteractive}
}
