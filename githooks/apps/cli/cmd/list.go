package cmd

import (
	"io"
	"path"
	cm "rycus86/githooks/common"
	"rycus86/githooks/git"
	"rycus86/githooks/hooks"
	strs "rycus86/githooks/strings"
	"strings"

	"github.com/pkg/math"
	"github.com/spf13/cobra"
)

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list [type]...",
	Short: "Lists the active hooks in the current repository.",
	Long: "Lists the active hooks in the current repository along with their state.\n" +
		"This command needs to be run at the root of a repository.\n\n" +
		"If 'type' is given, then it only lists the hooks for that trigger event.\n" +
		"The supported hooks are:\n\n" +
		strings.Join(strs.Map(hooks.ManagedHookNames, func(s string) string { return " • " + s }), "\n") +
		"\nThe value 'ns-path' is the namespaced path which is used for the ignore patterns.",
	PreRun: panicIfNotRangeArgs(0, 100),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			args = strs.MakeUnique(args)
			runList(args, true)
		} else {
			runList(hooks.ManagedHookNames, false)
		}
	}}

func runList(hookNames []string, warnNotFound bool) {
	repoDir, gitDir := assertRepoRoot(&settings)

	repoHooksDir := hooks.GetGithooksDir(repoDir)

	// Load checksum store
	checksums, err := hooks.GetChecksumStorage(settings.GitX, gitDir)
	log.AssertNoErrorF(err, "Errors while loading checksum store.")
	log.DebugF("%s", checksums.Summary())

	// Load ignore patterns
	ignores, err := hooks.GetIgnorePatterns(repoHooksDir, gitDir, hookNames)
	log.AssertNoErrorF(err, "Errors while loading ignore patterns.")
	log.DebugF("HooksDir ignore patterns: '%q'.", ignores.HooksDir)
	log.DebugF("User ignore patterns: '%q'.", ignores.User)

	// Load all shared hooks
	shared := sharedHooks{}
	shared.repository, err = hooks.LoadRepoSharedHooks(settings.InstallDir, repoDir)
	log.AssertNoErrorF(err, "Could not load repository shared hooks.")
	shared.local, err = hooks.LoadConfigSharedHooks(settings.InstallDir, settings.GitX, git.LocalScope)
	log.AssertNoErrorF(err, "Could not load local shared hooks.")
	shared.global, err = hooks.LoadConfigSharedHooks(settings.InstallDir, settings.GitX, git.GlobalScope)
	log.AssertNoErrorF(err, "Could not load global shared hooks.")

	pendingShared := filterPendingSharedHooks(&shared)

	isTrusted, _ := hooks.IsRepoTrusted(settings.GitX, repoDir)
	isDisabled := hooks.IsGithooksDisabled(settings.GitX, true)

	state := listHookState{
		checksums:          &checksums,
		ignores:            &ignores,
		isRepoTrusted:      isTrusted,
		isGithooksDisabled: isDisabled,
		sharedIgnores:      make(ignoresPerHookDir, 10),
		paddingMax:         60} //nolint: gomnd

	total := 0
	for _, hookName := range hookNames {

		list, count := listHooksForName(
			hookName,
			gitDir,
			repoHooksDir,
			&shared,
			&state)

		if count != 0 {
			log.InfoF("Hook: '%s' [%v] :%s", hookName, count, list)
		}

		total += count
	}

	printPendingShared(&pendingShared)

	log.InfoF("Total listed hooks: '%v'.", total)
}

type ignoresPerHookDir = map[string]*hooks.HookIgnorePatterns

type listHookState struct {
	checksums *hooks.ChecksumStore
	ignores   *hooks.RepoIgnorePatterns

	isRepoTrusted      bool
	isGithooksDisabled bool

	sharedIgnores ignoresPerHookDir
	paddingMax    int
}

const (
	sharedRepoCategory   = "shared:repo"
	sharedLocalCategory  = "shared:local"
	sharedGlobalCategory = "shared:global"
)

type sharedHooks struct {
	repository []hooks.SharedHook
	local      []hooks.SharedHook
	global     []hooks.SharedHook
}

func filterPendingSharedHooks(shared *sharedHooks) (pending sharedHooks) {

	// Filter out pending shared hooks.
	filter := func(shHooks []hooks.SharedHook) (res []hooks.SharedHook, pending []hooks.SharedHook) {
		res = make([]hooks.SharedHook, 0, len(shHooks))
		for _, sh := range shHooks {
			if cm.IsDirectory(sh.RepositoryDir) {
				res = append(res, sh)
			} else {
				pending = append(pending, sh)
			}
		}

		return
	}

	shared.repository, pending.repository = filter(shared.repository)
	shared.local, pending.local = filter(shared.local)
	shared.global, pending.global = filter(shared.global)

	return
}

func printPendingShared(shared *sharedHooks) {

	count := len(shared.repository) + len(shared.local) + len(shared.global)
	if count == 0 {
		return
	}

	var sb strings.Builder

	listPending := func(shHooks []hooks.SharedHook, indent string, category string) {
		for _, sh := range shHooks {
			_, err := strs.FmtW(&sb,
				"\n%s%s '%s' state: ['pending'], type: '%s'", indent, ListItemLiteral, sh.OriginalURL, category)
			cm.AssertNoErrorPanic(err, "Could not write pending hooks.")
		}
	}

	indent := " "
	listPending(shared.repository, indent, sharedRepoCategory)
	listPending(shared.local, indent, sharedLocalCategory)
	listPending(shared.global, indent, sharedGlobalCategory)

	log.InfoF("Pending shared hooks [%v] :%s", count, sb.String())
}

func listHooksForName(
	hookName string,
	gitDir string,
	repoHooksDir string,
	shared *sharedHooks,
	state *listHookState) (string, int) {

	var sb strings.Builder

	printHooks := func(hooks []hooks.Hook, title string, category string) {
		if len(hooks) == 0 {
			return
		}

		padding := findPaddingListHooks(hooks, state.paddingMax)
		_, err := strs.FmtW(&sb, "\n %s", title)
		cm.AssertNoErrorPanicF(err, "Could not write hook state.")

		for _, hook := range hooks {
			sb.WriteString("\n")
			formatHookState(&sb, hook, category, state.isGithooksDisabled, padding, "  ")
		}
	}

	listShared := func(sharedHooks []hooks.SharedHook, title string, category string) (count int) {

		for _, shHook := range sharedHooks {
			allHooks := getAllHooksIn(shHook.RepositoryDir, hookName, state, true, false)
			// @todo remove this as soon as possible
			allHooks = append(allHooks,
				getAllHooksIn(hooks.GetGithooksDir(shHook.RepositoryDir), hookName, state, true, false)...)

			printHooks(allHooks, strs.Fmt(title, shHook.OriginalURL), category)
			count += len(allHooks)
		}

		return count
	}

	// List replaced hooks (normally only one)
	replacedHooks := getAllHooksIn(path.Join(gitDir, "hooks"), hookName, state, false, true)
	printHooks(replacedHooks, "Replaced:", "replaced")

	// List repository hooks
	repoHooks := getAllHooksIn(repoHooksDir, hookName, state, false, false)
	printHooks(repoHooks, "Repository:", "repo")

	// List all shared hooks
	sharedCount :=
		listShared(shared.repository, "Shared: '%s'", sharedRepoCategory) +
			listShared(shared.local, "Shared: '%s'", sharedLocalCategory) +
			listShared(shared.global, "Shared: '%s'", sharedGlobalCategory)

	return sb.String(), len(replacedHooks) + len(repoHooks) + sharedCount
}

func findPaddingListHooks(hooks []hooks.Hook, maxPadding int) int {
	const addChars = 3
	max := 0
	for _, hook := range hooks {
		max = math.MaxInt(len(path.Base(hook.Path))+addChars, max)
	}

	return math.MinInt(max, maxPadding)
}

func getAllHooksIn(
	hooksDir string,
	hookName string,
	state *listHookState,
	addInternalIgnores bool,
	isReplacedHook bool) []hooks.Hook {

	isTrusted := func(hookPath string) (bool, string) {
		if state.isRepoTrusted {
			return true, ""
		}

		trusted, sha, e := state.checksums.IsTrusted(hookPath)
		log.AssertNoErrorF(e, "Could not check trust status '%s'.", hookPath)

		return trusted, sha
	}

	// Cache repository ignores
	hookDirIgnores := state.sharedIgnores[hooksDir]
	if hookDirIgnores == nil && addInternalIgnores {
		var e error
		igns, e := hooks.GetHookIgnorePatternsHookDir(hooksDir, []string{hookName})
		log.AssertNoErrorF(e, "Could not get worktree ignores in '%s'.", hooksDir)
		state.sharedIgnores[hooksDir] = &igns
		hookDirIgnores = &igns
	}

	isIgnored := func(namespacePath string) bool {
		ignored, byUser := state.ignores.IsIgnored(namespacePath)

		if isReplacedHook {
			return ignored && byUser // Replaced hooks can only be ignored by the user.
		} else if hookDirIgnores != nil {
			return ignored || hookDirIgnores.IsIgnored(namespacePath)
		}

		return ignored
	}

	hookNamespace, err := hooks.GetHooksNamespace(hooksDir)
	log.AssertNoErrorPanicF(err, "Could not get hook namespace in '%s'", hooksDir)

	if isReplacedHook {
		hookName = hooks.GetHookReplacementFileName(hookName)
		hookNamespace = "" // @todo Introduce namespacing here! use: "hooks"
	}

	allHooks, err := hooks.GetAllHooksIn(hooksDir, hookName, hookNamespace, isIgnored, isTrusted)
	log.AssertNoErrorPanicF(err, "Errors while collecting hooks in '%s'.", hooksDir)

	return allHooks
}

func formatHookState(
	w io.Writer,
	hook hooks.Hook,
	categeory string,
	isGithooksDisabled bool,
	padding int,
	indent string) {

	hooksFmt := strs.Fmt("%s%s %%-%vs : ", indent, ListItemLiteral, padding)
	const stateFmt = " state: ['%s', '%s']"
	const disabledStateFmt = " state: ['disabled']"
	const categeoryFmt = ", type: '%s'"
	const namespaceFmt = ", ns-path: '%s'"

	hookPath := strs.Fmt("'%s'", path.Base(hook.Path))

	if isGithooksDisabled {
		_, err := strs.FmtW(w,
			hooksFmt+disabledStateFmt+categeoryFmt+namespaceFmt,
			hookPath, categeory, hook.NamespacePath)

		cm.AssertNoErrorPanicF(err, "Could not write hook state.")

		return
	}

	active := "active" // nolint: goconst
	trusted := "trusted"

	if !hook.Active {
		active = "ignored"
	}

	if !hook.Trusted {
		trusted = "untrusted"
	}

	_, err := strs.FmtW(w,
		hooksFmt+stateFmt+categeoryFmt+namespaceFmt,
		hookPath, active, trusted, categeory, hook.NamespacePath)

	cm.AssertNoErrorPanicF(err, "Could not write hook state.")
}

func init() { // nolint: gochecknoinits
	rootCmd.AddCommand(setCommandDefaults(listCmd))
}
