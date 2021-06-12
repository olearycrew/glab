package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mgutz/ansi"

	surveyCore "github.com/AlecAivazis/survey/v2/core"
	"github.com/profclems/glab/commands"
	"github.com/profclems/glab/commands/alias/expand"
	"github.com/profclems/glab/commands/cmdutils"
	"github.com/profclems/glab/commands/help"
	"github.com/profclems/glab/commands/update"
	"github.com/profclems/glab/internal/config"
	"github.com/profclems/glab/internal/glinstance"
	"github.com/profclems/glab/internal/run"
	"github.com/profclems/glab/pkg/tableprinter"
	"github.com/profclems/glab/pkg/utils"

	"github.com/spf13/cobra"
)

// version is set dynamically at build
var version = "DEV"

// build is set dynamically at build
var build string

// debug is set dynamically at build and can be overridden by
// the configuration file or environment variable
// sets to "true" or "false" or "1" or "0" as string
var debugMode = "false"

// debug is parsed boolean of debugMode
var debug bool

func main() {
	debug = debugMode == "true" || debugMode == "1"

	cmdFactory := cmdutils.NewFactory()

	maybeOverrideDefaultHost(cmdFactory)

	if !cmdFactory.IO.ColorEnabled() {
		surveyCore.DisableColor = true
	} else {
		// Override survey's choice of color for default values
		// For default values for e.g. `Input` prompts, Survey uses the literal "white" color,
		// which makes no sense on dark terminals and is literally invisible on light backgrounds.
		// This overrides Survey to output a gray color for 256-color terminals and "default" for basic terminals.
		surveyCore.TemplateFuncsWithColor["color"] = func(style string) string {
			switch style {
			case "white":
				if cmdFactory.IO.Is256ColorSupported() {
					return fmt.Sprintf("\x1b[%d;5;%dm", 38, 242)
				}
				return ansi.ColorCode("default")
			default:
				return ansi.ColorCode(style)
			}
		}
	}

	rootCmd := commands.NewCmdRoot(cmdFactory, version, build)

	cfg, err := cmdFactory.Config()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read configuration:  %s\n", err)
		os.Exit(2)
	}

	// Set Debug mode
	debugMode, _ = cfg.Get("", "debug")
	debug = debugMode == "true" || debugMode == "1"

	if pager, _ := cfg.Get("", "glab_pager"); pager != "" {
		cmdFactory.IO.SetPager(pager)
	}

	if promptDisabled, _ := cfg.Get("", "no_prompt"); promptDisabled != "" {
		cmdFactory.IO.SetPrompt(promptDisabled)
	}

	var expandedArgs []string
	if len(os.Args) > 0 {
		expandedArgs = os.Args[1:]
	}

	cmd, _, err := rootCmd.Traverse(expandedArgs)
	if err != nil || cmd == rootCmd {
		originalArgs := expandedArgs
		isShell := false
		expandedArgs, isShell, err = expand.ExpandAlias(cfg, os.Args, nil)
		if err != nil {
			fmt.Fprintf(os.Stdout, "Failed to process alias: %s\n", err)
			os.Exit(2)
		}

		if debug {
			fmt.Printf("%v -> %v\n", originalArgs, expandedArgs)
		}

		if isShell {
			externalCmd := exec.Command(expandedArgs[0], expandedArgs[1:]...)
			externalCmd.Stderr = os.Stderr
			externalCmd.Stdout = os.Stdout
			externalCmd.Stdin = os.Stdin
			preparedCmd := run.PrepareCmd(externalCmd)

			err = preparedCmd.Run()
			if err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					os.Exit(ee.ExitCode())
				}

				fmt.Fprintf(os.Stdout, "failed to run external command: %s", err)
				os.Exit(3)
			}

			os.Exit(0)
		}
	}

	// Override the default column separator of tableprinter
	tableprinter.SetSeparator("  ")
	// Override the default terminal width of tableprinter
	tableprinter.SetTerminalWidth(cmdFactory.IO.TerminalWidth())

	rootCmd.SetArgs(expandedArgs)

	if cmd, err := rootCmd.ExecuteC(); err != nil {
		printError(os.Stderr, err, cmd, debug)
		cmd.Print("\n")
		os.Exit(1)
	}

	if help.HasFailed() {
		os.Exit(1)
	}

	checkUpdate, _ := cfg.Get("", "check_update")
	if checkUpdate, err := strconv.ParseBool(checkUpdate); err == nil && checkUpdate {
		err = update.CheckUpdate(cmdFactory.IO, version, true)
		if err != nil && debug {
			printError(os.Stderr, err, rootCmd, debug)
		}
	}
}

func printError(out io.Writer, err error, cmd *cobra.Command, debug bool) {
	if err == cmdutils.SilentError {
		return
	}

	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		_, _ = fmt.Fprintf(out, "error connecting to %s\n", dnsError.Name)
		if debug {
			_, _ = fmt.Fprintln(out, dnsError)
		}
		_, _ = fmt.Fprintln(out, "check your internet connection or status.gitlab.com or 'Run sudo gitlab-ctl status' on your server if self-hosted")
		return
	}
	_, _ = fmt.Fprintln(out, err)

	var flagError *cmdutils.FlagError
	if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {
		if !strings.HasSuffix(err.Error(), "\n") {
			_, _ = fmt.Fprintln(out)
		}
		_, _ = fmt.Fprintln(out, cmd.UsageString())
	}
}

func maybeOverrideDefaultHost(f *cmdutils.Factory) {
	baseRepo, err := f.BaseRepo()
	if err == nil {
		glinstance.OverrideDefault(baseRepo.RepoHost())
	}
	if glHostFromEnv := config.GetFromEnv("host"); glHostFromEnv != "" {
		if utils.IsValidURL(glHostFromEnv) {
			var protocol string
			glHostFromEnv, protocol = glinstance.StripHostProtocol(glHostFromEnv)
			glinstance.OverrideDefaultProtocol(protocol)
		}
		glinstance.OverrideDefault(glHostFromEnv)
	}
}
