package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"rycus86/githooks/build"
	"rycus86/githooks/builder"
	cm "rycus86/githooks/common"
	"rycus86/githooks/git"
	"rycus86/githooks/hooks"
	"rycus86/githooks/prompt"
	strs "rycus86/githooks/strings"
	"rycus86/githooks/updates"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// InstallSettings are the settings for the installer.
type InstallSettings struct {
	installDir string
	cloneDir   string

	promptCtx prompt.IContext

	hookTemplateDir string
}

var log cm.ILogContext
var args = Arguments{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "githooks-installer",
	Short: "Githooks installer application",
	Long: "Githooks installer application\n" +
		"See further information at https://github.com/rycus86/githooks/blob/master/README.md",
	Run: runInstall}

// ProxyWriterOut is solely used for the cobra logging.
type ProxyWriterOut struct {
	log cm.ILogContext
}

// ProxyWriterErr is solely used for the cobra logging.
type ProxyWriterErr struct {
	log cm.ILogContext
}

func (p *ProxyWriterOut) Write(s []byte) (int, error) {
	return p.log.GetInfoWriter().Write([]byte(p.log.ColorInfo(string(s))))
}

func (p *ProxyWriterErr) Write(s []byte) (int, error) {
	return p.log.GetErrorWriter().Write([]byte(p.log.ColorError(string(s))))
}

// Run adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Run() {
	cobra.OnInitialize(initArgs)

	rootCmd.SetOut(&ProxyWriterOut{log: log})
	rootCmd.SetErr(&ProxyWriterErr{log: log})
	rootCmd.Version = build.BuildVersion

	defineArguments(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initArgs() {

	viper.BindEnv("internalAutoUpdate", "GITHOOKS_INTERNAL_AUTO_UPDATE")

	config := viper.GetString("config")
	if strs.IsNotEmpty(config) {
		viper.SetConfigFile(config)
		err := viper.ReadInConfig()
		log.AssertNoErrorPanicF(err, "Could not read config file '%s'.", config)
	}

	err := viper.Unmarshal(&args)
	log.AssertNoErrorPanicF(err, "Could not unmarshal parameters.")
}

func writeArgs(file string, args *Arguments) {
	err := cm.StoreJSON(file, args)
	log.AssertNoErrorPanicF(err, "Could not write arguments to '%s'.", file)
}

func defineArguments(rootCmd *cobra.Command) {
	// Internal commands
	rootCmd.PersistentFlags().String("config", "",
		"JSON config according to the 'Arguments' struct.")
	rootCmd.MarkPersistentFlagDirname("config")
	rootCmd.PersistentFlags().MarkHidden("config")

	rootCmd.PersistentFlags().Bool("internal-auto-update", false,
		"Internal argument, do not use!") // @todo Remove this...

	// User commands
	rootCmd.PersistentFlags().Bool("dry-run", false,
		"Dry run the installation showing whats beeing done.")

	rootCmd.PersistentFlags().Bool(
		"non-interactive", false,
		"Run the installation non-interactively\n"+
			"without showing prompts.")

	rootCmd.PersistentFlags().Bool(
		"single", false,
		"Install Githooks in the active repository only.\n"+
			"This does not mean it won't install necessary\n"+
			"files into the installation directory.")

	rootCmd.PersistentFlags().Bool(
		"skip-install-into-existing", false,
		"Skip installation into existing repositories\n"+
			"defined by a search path.")

	rootCmd.PersistentFlags().String(
		"prefix", "",
		"Githooks installation prefix such that\n"+
			"'<prefix>/.githooks' will be the installation directory.")
	rootCmd.MarkPersistentFlagDirname("config")

	rootCmd.PersistentFlags().String(
		"template-dir", "",
		"The preferred template directory to use.")

	rootCmd.PersistentFlags().Bool(
		"only-server-hooks", false,
		"Only install and maintain server hooks.")

	rootCmd.PersistentFlags().Bool(
		"use-core-hookspath", false,
		"If the install mode 'core.hooksPath' should be used.")

	rootCmd.PersistentFlags().String(
		"clone-url", "",
		"The clone url from which Githooks should clone\n"+
			"and install itself.")

	rootCmd.PersistentFlags().String(
		"clone-branch", "",
		"The clone branch from which Githooks should\n"+
			"clone and install itself.")

	rootCmd.PersistentFlags().Bool(
		"build-from-source", false,
		"If the binaries are built from source instead of\n"+
			"downloaded from the deploy url.")

	rootCmd.PersistentFlags().StringSlice(
		"build-flags", nil,
		"Build flags for building from source (get extended with defaults).")

	rootCmd.PersistentFlags().Bool(
		"stdin", false,
		"Use standard input to read prompt answers.")

	rootCmd.Args = cobra.NoArgs

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("internalAutoUpdate", rootCmd.PersistentFlags().Lookup("internal-auto-update")) // @todo Remove this...
	viper.BindPFlag("dryRun", rootCmd.PersistentFlags().Lookup("dry-run"))
	viper.BindPFlag("nonInteractive", rootCmd.PersistentFlags().Lookup("non-interactive"))
	viper.BindPFlag("singleInstall", rootCmd.PersistentFlags().Lookup("single"))
	viper.BindPFlag("skipInstallIntoExisting", rootCmd.PersistentFlags().Lookup("skip-install-into-existing"))
	viper.BindPFlag("onlyServerHooks", rootCmd.PersistentFlags().Lookup("only-server-hooks"))
	viper.BindPFlag("useCoreHooksPath", rootCmd.PersistentFlags().Lookup("use-core-hookspath"))
	viper.BindPFlag("cloneURL", rootCmd.PersistentFlags().Lookup("clone-url"))
	viper.BindPFlag("cloneBranch", rootCmd.PersistentFlags().Lookup("clone-branch"))
	viper.BindPFlag("buildFromSource", rootCmd.PersistentFlags().Lookup("build-from-source"))
	viper.BindPFlag("buildFlags", rootCmd.PersistentFlags().Lookup("build-flags"))
	viper.BindPFlag("installPrefix", rootCmd.PersistentFlags().Lookup("prefix"))
	viper.BindPFlag("templateDir", rootCmd.PersistentFlags().Lookup("template-dir"))
	viper.BindPFlag("useStdin", rootCmd.PersistentFlags().Lookup("stdin"))
}

func validateArgs(cmd *cobra.Command, args *Arguments) {

	// Check all parsed flags to not have empty value!
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		log.PanicIfF(f.Changed && strs.IsEmpty(f.Value.String()),
			"Flag '%s' needs an non-empty value.", f.Name)
	})

	log.PanicIfF(args.SingleInstall && args.UseCoreHooksPath,
		"Cannot use --single and --use-core-hookspath together. See `--help`.")
}

func setMainVariables(args *Arguments) InstallSettings {

	var promptCtx prompt.IContext
	var err error

	if !args.NonInteractive {
		promptCtx, err = prompt.CreateContext(log, &cm.ExecContext{}, nil, false, args.UseStdin)
		log.AssertNoErrorF(err, "Prompt setup failed -> using fallback.")
	}

	var installDir string
	// First check if we already have
	// an install directory set (from --prefix)
	if strs.IsNotEmpty(args.InstallPrefix) {
		var err error
		args.InstallPrefix, err = cm.ReplaceTilde(filepath.ToSlash(args.InstallPrefix))
		log.AssertNoErrorPanic(err, "Could not replace '~' character in path.")
		installDir = path.Join(args.InstallPrefix, ".githooks")

	} else {
		installDir = hooks.GetInstallDir()
		if !cm.IsDirectory(installDir) {
			log.WarnF("Install directory '%s' does not exist.\n"+
				"Setting to default '~/.githooks'.", installDir)
			installDir = ""
		}
	}

	if strs.IsEmpty(installDir) {
		installDir, err = homedir.Dir()
		cm.AssertNoErrorPanic(err, "Could not get home directory.")
		installDir = path.Join(filepath.ToSlash(installDir), hooks.HookDirName)
	}

	return InstallSettings{
		installDir: installDir,
		cloneDir:   hooks.GetReleaseCloneDir(installDir),
		promptCtx:  promptCtx}
}

func setInstallDir(installDir string) {
	log.AssertNoErrorPanic(hooks.SetInstallDir(installDir),
		"Could not set install dir '%s'", installDir)
}

func buildFromSource(
	args *Arguments,
	tempDir string,
	url string,
	branch string,
	commitSHA string) updates.Binaries {

	log.Info("Building binaries from source ...")

	// Clone another copy of the release clone into temporary directory
	log.InfoF("Clone to temporary build directory '%s'", tempDir)
	err := git.Clone(tempDir, url, branch, -1)
	log.AssertNoErrorPanicF(err, "Could not clone release branch into '%s'.", tempDir)

	// Checkout the remote commit sha
	log.InfoF("Checkout out commit '%s'", commitSHA[0:6])
	gitx := git.CtxC(tempDir)
	err = gitx.Check("checkout",
		"-b", "update-to-"+commitSHA[0:6],
		commitSHA)

	log.AssertNoErrorPanicF(err,
		"Could not checkout update commit '%s' in '%s'.",
		commitSHA, tempDir)

	tag, _ := gitx.Get("describe", "--tags", "--abbrev=6")
	log.InfoF("Building binaries at '%s'", tag)

	// Build the binaries.
	binPath, err := builder.Build(tempDir, args.BuildFlags)
	log.AssertNoErrorPanicF(err, "Could not build release branch in '%s'.", tempDir)

	bins, err := cm.GetAllFiles(binPath)
	log.AssertNoErrorPanicF(err, "Could not get files in path '%s'.", binPath)

	binaries := updates.Binaries{BinDir: binPath}
	strs.Map(bins, func(s string) string {
		if cm.IsExecutable(s) {
			if strings.Contains(s, "installer") {
				binaries.Installer = s
			} else {
				binaries.Others = append(binaries.Others, s)
			}
			binaries.All = append(binaries.All, s)
		}
		return s
	})

	log.InfoF(
		"Successfully built %v binaries:\n - %s",
		len(binaries.All),
		strings.Join(
			strs.Map(binaries.All,
				func(s string) string { return path.Base(s) }),
			"\n - "))

	log.PanicIf(
		len(binaries.All) == 0 ||
			strs.IsEmpty(binaries.Installer),
		"No binaries found in '%s'", binPath)

	return binaries
}

func downloadBinaries(settings *InstallSettings, tempDir string, status updates.ReleaseStatus) updates.Binaries {
	log.Panic("Not implemented")
	return updates.Binaries{}
}

func prepareDispatch(settings *InstallSettings, args *Arguments) {

	var status updates.ReleaseStatus
	var err error

	if args.InternalAutoUpdate {

		status, err = updates.GetStatus(settings.cloneDir, true)
		log.AssertNoErrorPanic(err,
			"Could not get status of release clone '%s'",
			settings.cloneDir)

	} else {

		status, err = updates.FetchUpdates(
			settings.cloneDir,
			args.CloneURL,
			args.CloneBranch,
			true, updates.RecloneOnWrongRemote)

		log.AssertNoErrorPanicF(err,
			"Could not assert release clone '%s' existing",
			settings.cloneDir)
	}

	tempDir, err := ioutil.TempDir(os.TempDir(), "githooks-update")
	log.AssertNoErrorPanic(err, "Can not create temporary update dir in '%s'", os.TempDir())
	defer os.RemoveAll(tempDir)

	updateSettings := updates.GetSettings()

	binaries := updates.Binaries{}
	if args.BuildFromSource || updateSettings.DoBuildFromSource {
		binaries = buildFromSource(
			args, tempDir,
			status.RemoteURL, status.Branch, status.RemoteCommitSHA)
	} else {
		_ = downloadBinaries(settings, tempDir, status)
	}

	updateTo := ""
	if status.LocalCommitSHA != status.RemoteCommitSHA {
		updateTo = status.RemoteCommitSHA
	}

	runInstaller(binaries.Installer, args, tempDir, updateTo, binaries.All)
}

func runInstaller(installer string, args *Arguments, tempDir string, updateTo string, binaries []string) {
	log.Info("Dispatching to build installer ...")

	// Set variables...
	args.InternalPostUpdate = true
	args.InternalUpdateTo = updateTo
	args.InternalBinaries = binaries

	file, err := ioutil.TempFile(tempDir, "*install-config.json")
	log.AssertNoErrorPanicF(err, "Could not create temporary file in '%s'.", tempDir)
	defer os.Remove(file.Name())

	// Write the config ...
	writeArgs(file.Name(), args)

	// Run the installer binary
	err = cm.RunExecutable(
		&cm.ExecContext{},
		&cm.Executable{Path: installer},
		true,
		"--config", file.Name())

	log.AssertNoErrorPanic(err, "Running installer failed.")
}

func checkTemplateDir(targetDir string) string {
	if strs.IsEmpty(targetDir) {
		return ""
	}

	if cm.IsWritable(targetDir) {
		return targetDir
	}

	targetDir, err := cm.ReplaceTilde(targetDir)
	log.AssertNoErrorPanicF(err,
		"Could not replace tilde '~' in '%s'.", targetDir)

	if cm.IsWritable(targetDir) {
		return targetDir
	}

	return ""
}

// findGitHookTemplates returns the Git hook template directory
// and optional a Git template dir which is only set in case of
// not using the core.hooksPath method.
func findGitHookTemplates(
	installDir string,
	useCoreHooksPath bool,
	nonInteractive bool,
	promptCtx prompt.IContext) (string, string) {

	installUsesCoreHooksPath := git.Ctx().GetConfig("githooks.useCoreHooksPath", git.GlobalScope)
	haveInstallation := strs.IsNotEmpty(installUsesCoreHooksPath)

	// 1. Try setup from environment variables
	gitTempDir, exists := os.LookupEnv("GIT_TEMPLATE_DIR")
	if exists {
		templateDir := checkTemplateDir(gitTempDir)

		if strs.IsNotEmpty(templateDir) {
			return path.Join(templateDir, "hooks"), ""
		}
	}

	// 2. Try setup from git config
	if useCoreHooksPath || installUsesCoreHooksPath == "true" {
		hooksTemplateDir := checkTemplateDir(
			git.Ctx().GetConfig("core.hooksPath", git.GlobalScope))

		if strs.IsNotEmpty(hooksTemplateDir) {
			return hooksTemplateDir, ""
		}
	} else {
		templateDir := checkTemplateDir(
			git.Ctx().GetConfig("init.templateDir", git.GlobalScope))

		if strs.IsNotEmpty(templateDir) {
			return path.Join(templateDir, "hooks"), ""
		}
	}

	// 3. Try setup from the default location
	hooksTemplateDir := checkTemplateDir(path.Join(git.GetDefaultTemplateDir(), "hooks"))
	if strs.IsNotEmpty(hooksTemplateDir) {
		return hooksTemplateDir, ""
	}

	// If we have an installation, and have not found
	// the template folder by now...
	log.PanicIfF(haveInstallation,
		"Your installation is corrupt.\n"+
			"The global value 'githooks.useCoreHooksPath = %v'\n"+
			"is set but the corresponding hook templates directory\n"+
			"is not found.")

	// 4. Try setup new folder if running non-interactively
	// and no folder is found by now
	if nonInteractive {
		templateDir := setupNewTemplateDir(installDir, nil)
		return path.Join(templateDir, "hooks"), templateDir
	}

	// 5. Try to search for it on disk
	answer, err := promptCtx.ShowPromptOptions(
		"Could not find the Git hook template directory.\n"+
			"Do you want to search for it?",
		"(yes, No)",
		"y/N",
		"Yes", "No")
	log.AssertNoErrorF(err, "Could not show prompt.")

	if answer == "y" {

		templateDir := searchTemplateDirOnDisk(promptCtx)

		if strs.IsNotEmpty(templateDir) {

			if useCoreHooksPath {
				return path.Join(templateDir, "hooks"), ""
			}

			// If we dont use core.hooksPath, we ask
			// if the user wants to continue setting this as
			// 'init.templateDir'.
			answer, err := promptCtx.ShowPromptOptions(
				"Do you want to set this up as the Git template\n"+
					"directory (e.g setting 'init.templateDir')\n"+
					"for future use?",
				"(yes, No (abort))",
				"y/N",
				"Yes", "No (abort)")
			log.AssertNoErrorF(err, "Could not show prompt.")

			log.PanicIf(answer != "y",
				"Could not determine Git hook",
				"templates directory. -> Abort.")

			return path.Join(templateDir, "hooks"), templateDir
		}
	}

	// 6. Set up as new
	answer, err = promptCtx.ShowPromptOptions(
		"Do you want to set up a new Git templates folder?",
		"(yes, No)",
		"y/N",
		"Yes", "No")
	log.AssertNoErrorF(err, "Could not show prompt.")

	if answer == "y" {
		templateDir := setupNewTemplateDir(installDir, promptCtx)
		return path.Join(templateDir, "hooks"), templateDir
	}

	return "", ""
}

func searchPreCommitFile(startDirs []string, promptCtx prompt.IContext) string {

	for _, dir := range startDirs {

		log.InfoF("Searching for potential locations in '%s'...", dir)

		spinner := cm.GetProgressBar(log, "Searching ...", -1)

		// Search in go routine...
		matchCh := make(chan []string, 1)
		go func() {
			matches, err := cm.Glob(path.Join(dir,
				"**/templates/hooks/pre-commit.sample"))
			if err == nil {
				matchCh <- matches
			}
		}()

		var matches []string
		running := true

		spinnerT := time.NewTicker(10 * time.Millisecond)
		stillSearching := time.After(10 * time.Second)
		timeout := time.After(300 * time.Second)

		for running {
			select {
			case matches = <-matchCh:
				running = false
				if spinner != nil {
					spinner.Clear()
				}

			case <-stillSearching:
				if spinner == nil {
					log.Info("Still searching ...")
				} else {
					spinner.Describe("Still searching ...")
				}
			case <-spinnerT.C:
				if spinner != nil {
					spinner.Add(1)
				}
			case <-timeout:
				log.Warn("Searching took too long -> timed out.")
				running = false
			}
		}

		for _, match := range matches {

			templateDir := path.Dir(path.Dir(filepath.ToSlash(match)))

			answer, err := promptCtx.ShowPromptOptions(
				strs.Fmt("--> Is it '%s'", templateDir),
				"(yes, No)",
				"y/N",
				"Yes", "No")
			log.AssertNoErrorF(err, "Could not show prompt.")

			if answer == "y" {
				return templateDir
			}
		}
	}

	return ""
}

func searchTemplateDirOnDisk(promptCtx prompt.IContext) string {

	first, second := GetDefaultTemplateSearchDir()

	templateDir := searchPreCommitFile(first, promptCtx)

	if strs.IsEmpty(templateDir) {

		answer, err := promptCtx.ShowPromptOptions(
			"Git hook template directory not found\n"+
				"Do you want to keep searching?",
			"(yes, No)",
			"y/N",
			"Yes", "No")

		log.AssertNoErrorF(err, "Could not show prompt.")

		if answer == "y" {
			templateDir = searchPreCommitFile(second, promptCtx)
		}
	}

	return templateDir
}

func setupNewTemplateDir(installDir string, promptCtx prompt.IContext) string {
	templateDir := path.Join(installDir, "templates")

	if promptCtx != nil {
		var err error
		templateDir, err = promptCtx.ShowPrompt(
			"Enter the target folder", templateDir, false)
		log.AssertNoErrorF(err, "Could not show prompt.")
	}

	templateDir, err := cm.ReplaceTilde(templateDir)
	log.AssertNoErrorPanicF(err, "Could not replace tilde '~' in '%s'.", templateDir)

	return templateDir
}

func getTargetTemplateDir(
	installDir string,
	templateDir string,
	useCoreHooksPath bool,
	nonInteractive bool,
	dryRun bool,
	promptCtx prompt.IContext) (hookTemplateDir string) {

	if strs.IsEmpty(templateDir) {
		// Automatically find a template directory.
		hookTemplateDir, templateDir = findGitHookTemplates(
			installDir,
			useCoreHooksPath,
			nonInteractive,
			promptCtx)

		log.PanicIfF(strs.IsEmpty(hookTemplateDir),
			"Could not determine Git hook template directory.")
	} else {
		// The user provided a template directory, check it and
		// add `hooks` which is needed.
		log.PanicIfF(!cm.IsDirectory(templateDir),
			"Given template dir '%s' does not exist.", templateDir)
		hookTemplateDir = path.Join(templateDir, "hooks")
	}

	err := os.MkdirAll(hookTemplateDir, 0775)
	log.AssertNoErrorPanicF(err,
		"Could not assert directory '%s' exists",
		hookTemplateDir)

	// Set the global Git configuration
	if useCoreHooksPath {
		setGithooksDirectory(true, hookTemplateDir, dryRun)
	} else {
		setGithooksDirectory(false, templateDir, dryRun)
	}

	return
}

func setGithooksDirectory(useCoreHooksPath bool, directory string, dryRun bool) {
	gitx := git.Ctx()

	prefix := "Setting"
	if dryRun {
		prefix = "[dry run] Would set"
	}

	if useCoreHooksPath {

		log.InfoF("%s 'core.hooksPath' to '%s'.", prefix, directory)

		if !dryRun {
			err := gitx.SetConfig("githooks.useCoreHooksPath", true, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")

			err = gitx.SetConfig("githooks.pathForUseCoreHooksPath", directory, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")

			err = gitx.SetConfig("core.hooksPath", directory, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")
		}

		// Warnings:
		// Check if hooks might not run...
		tD := gitx.GetConfig("init.templateDir", git.GlobalScope)
		msg := ""
		if strs.IsNotEmpty(tD) && cm.IsDirectory(path.Join(tD, "hooks")) {
			d := path.Join(tD, "hooks")
			files, err := cm.GetAllFiles(d)
			log.AssertNoErrorPanicF(err, "Could not get files in '%s'.", d)

			if len(files) > 0 {
				msg = strs.Fmt(
					"The 'init.templateDir' setting is currently set to\n"+
						"'%s'\n"+
						"and contains '%v' potential hooks.\n", tD, len(files))
			}
		}

		tDEnv := os.Getenv("GIT_TEMPLATE_DIR")
		if strs.IsNotEmpty(tDEnv) && cm.IsDirectory(path.Join(tDEnv, "hooks")) {
			d := path.Join(tDEnv, "hooks")
			files, err := cm.GetAllFiles(d)
			log.AssertNoErrorPanicF(err, "Could not get files in '%s'.", d)

			if len(files) > 0 {
				msg += strs.Fmt(
					"The environment variable 'GIT_TEMPLATE_DIR' is currently set to\n"+
						"'%s'\n"+
						"and contains '%v' potential hooks.\n", tDEnv, len(files))
			}
		}

		log.WarnIf(strs.IsNotEmpty(msg),
			msg+
				"These hooks might get installed but\n"+
				"ignored because 'core.hooksPath' is also set.\n"+
				"It is recommended to either remove the files or run\n"+
				"the Githooks installation without the '--use-core-hookspath'\n"+
				"parameter.")

	} else {

		if !dryRun {
			err := gitx.SetConfig("githooks.useCoreHooksPath", false, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")
		}

		if strs.IsNotEmpty(directory) {
			log.InfoF("%s 'init.templateDir' to '%s'.", prefix, directory)

			if !dryRun {
				err := gitx.SetConfig("init.templateDir", directory, git.GlobalScope)
				log.AssertNoErrorPanic(err, "Could not set Git config value.")
			}
		}

		// Warnings:
		// Check if hooks might not run..
		hP := gitx.GetConfig("core.hooksPath", git.GlobalScope)
		log.WarnIfF(strs.IsNotEmpty(hP),
			"The 'core.hooksPath' setting is currently set to\n"+
				"'%s'\n"+
				"This could mean that Githooks hooks will be ignored\n"+
				"Either unset 'core.hooksPath' or run the Githooks\n"+
				"installation with the '--use-core-hookspath' parameter.",
			hP)

	}
}

func setupHookTemplates(
	hookTemplateDir string,
	cloneDir string,
	onlyServerHooks bool,
	dryRun bool) {

	if dryRun {
		log.InfoF("[dry run] Would install Git hook templates into '%s'.",
			hookTemplateDir)
		return
	}

	log.InfoF("Installing Git hook templates into '%s'.",
		hookTemplateDir)

	var hookNames []string
	if onlyServerHooks {
		hookNames = managedServerHookNames
	} else {
		hookNames = managedHookNames
	}

	for _, hookName := range hookNames {
		dest := path.Join(hookTemplateDir, hookName)

		// Check there is already a Git hook in place and replace it.
		if cm.IsFile(dest) {
			isTemplate, err := hooks.IsRunWrapper(dest)
			log.AssertNoErrorF(err, "Could not detect if '%s' is a custom Git hook")

			if !isTemplate {
				log.InfoF("Saving existing Git hook '%s'.", dest)

				newDest := path.Join(hookTemplateDir, hooks.GetRunWrapperReplacementName(hookName))

				err := os.Rename(dest, newDest)
				log.AssertNoErrorPanicF(err, "Could not rename file '%s' to '%s'.", dest, newDest)

			}
		}

		err := hooks.WriteRunWrapper(dest)
		log.AssertNoErrorPanicF(err, "Could not write run wrapper to '%s'.", dest)
	}
}

func installBinaries(installDir string, cloneDir string, binaries []string, dryRun bool) {

	binDir := hooks.GetBinaryDir(installDir)

	err := os.MkdirAll(binDir, 0775)
	log.AssertNoErrorPanicF(err, "Could not create binary dir '%s'.", binDir)

	msg := strs.Map(binaries, func(s string) string { return strs.Fmt(" - '%s'", s) })
	if dryRun {
		log.InfoF("[dry run] Would install binaries:\n%s\n"+"to '%s'.", msg)
		return
	}

	log.InfoF("Installing binaries:\n%s\n"+"to '%s'.", strings.Join(msg, "\n"), binDir)

	if hooks.InstallLegacyBinaries {
		binaries = append(binaries, path.Join(cloneDir, "cli.sh"))
	}

	for _, binary := range binaries {
		dest := path.Join(binDir, path.Base(binary))
		err := cm.MoveFileWithBackup(binary, dest)
		log.AssertNoErrorPanicF(err,
			"Could not move file '%s' to '%s'.", binary, dest)
	}

	// Set CLI executable alias.
	cliTool := hooks.GetCLIExecutable(installDir)
	err = hooks.SetCLIExecutableAlias(cliTool)
	log.AssertNoErrorPanicF(err,
		"Could not set Git config 'alias.hooks' to '%s'.", cliTool)

	if hooks.InstallLegacyBinaries {
		err = cm.MakeExecutbale(cliTool)
		log.AssertNoErrorPanicF(err, "Making file '%s' executable failed.", cliTool)
	}

	// Set runner executable alias.
	runner := hooks.GetRunnerExecutable(installDir)
	err = hooks.SetRunnerExecutableAlias(runner)
	log.AssertNoErrorPanic(err,
		"Could not set runner executable alias '%s'.", runner)
}

func setupAutomaticUpdate(nonInteractive bool, dryRun bool, promptCtx prompt.IContext) {
	gitx := git.Ctx()
	currentSetting := gitx.GetConfig("githooks.autoupdate.enable", git.GlobalScope)
	promptMsg := ""

	if currentSetting == "true" {
		// Already enabled.
		return
	} else if strs.IsEmpty(currentSetting) {
		promptMsg = "Would you like to enable automatic update checks,\ndone once a day after a commit?"

	} else {
		log.Info("Automatic update checks are currently disabled.")
		if nonInteractive {
			return
		}
		promptMsg = "Would you like to re-enable them,\ndone once a day after a commit?"
	}

	activate := false

	if nonInteractive {
		activate = true
	} else {
		answer, err := promptCtx.ShowPromptOptions(
			promptMsg,
			"(Yes, no)",
			"Y/n", "Yes", "No")
		log.AssertNoErrorF(err, "Could not show prompt.")

		activate = answer == "y"
	}

	if activate {
		if dryRun {
			log.Info("[dry run] Would enable automatic update checks.")
		} else {

			if err := gitx.SetConfig(
				"githooks.autoupdate.enabled", true, git.GlobalScope); err == nil {

				log.Info("Automatic update checks are now enabled.")
			} else {
				log.Error("Failed to enable automatic update checks.")
			}

		}
	} else {
		log.Info(
			"If you change your mind in the future, you can enable it by running:",
			"  $ git hooks update enable")
	}
}

func runUpdate(settings *InstallSettings, args *Arguments) {
	log.InfoF("Running update to version '%s' ...", build.BuildVersion)

	if args.NonInteractive {
		// disable the prompt context,
		// no prompt must be shown in this mode
		// if we do -> pandic...
		settings.promptCtx = nil
	}

	settings.hookTemplateDir = getTargetTemplateDir(
		settings.installDir,
		args.TemplateDir,
		args.UseCoreHooksPath,
		args.NonInteractive,
		args.DryRun,
		settings.promptCtx)

	installBinaries(
		settings.installDir,
		settings.cloneDir,
		args.InternalBinaries,
		args.DryRun)

	if !args.InternalAutoUpdate {
		setupAutomaticUpdate(args.NonInteractive, args.DryRun, settings.promptCtx)
	}

	setupHookTemplates(
		settings.hookTemplateDir,
		settings.cloneDir,
		args.OnlyServerHooks,
		args.DryRun)
}

func runInstall(cmd *cobra.Command, auxArgs []string) {

	log.DebugF("Arguments: %+v", args)
	validateArgs(cmd, &args)

	settings := setMainVariables(&args)

	if !args.DryRun {
		setInstallDir(settings.installDir)
	}

	if !args.InternalPostUpdate {
		prepareDispatch(&settings, &args)
	} else {
		runUpdate(&settings, &args)
	}

}

func main() {

	cwd, err := os.Getwd()
	cm.AssertNoErrorPanic(err, "Could not get current working dir.")
	cwd = filepath.ToSlash(cwd)

	log, err = cm.CreateLogContext(cm.IsRunInDocker)
	cm.AssertOrPanic(err == nil, "Could not create log")

	log.InfoF("Installer [version: %s]", build.BuildVersion)

	var exitCode int
	defer func() { os.Exit(exitCode) }()

	// Handle all panics and report the error
	defer func() {
		r := recover()
		if hooks.HandleCLIErrors(r, cwd, log) {
			exitCode = 1
		}
	}()

	Run()
}
