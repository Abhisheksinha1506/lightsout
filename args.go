package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// command is what lightsout was asked to do by its arguments.
type command int

const (
	runQuit command = iota
	showVersion
	showHelp
	keepApps
	unkeepApps
	printCompletion
)

const (
	// defaultTimeout is how long lightsout waits for each app to quit.
	defaultTimeout = 10 * time.Second
	maxTimeout     = 300 * time.Second
)

// options are lightsout's parsed command-line arguments.
type options struct {
	cmd     command
	apps    []string // apps to keep (runQuit, keepApps, unkeepApps) or the shell (printCompletion)
	dryRun  bool
	yes     bool
	timeout time.Duration
}

const usage = `Usage:
  lightsout [options] [app ...]   Quit every open app except these
  lightsout keep app ...          Always keep these apps open
  lightsout unkeep app ...        Stop always keeping these apps open
  lightsout completion SHELL      Print a completion script (bash, zsh, fish, powershell)

Asks every open app to quit, except the app you run lightsout from and the
apps named as arguments or in ~/.config/lightsout/config.yml. Nothing is
force quit.

Options:
  -n, --dry-run        Show what would be quit, without quitting anything
  -y, --yes            Don't ask which app you're running from; stop if it
                       can't be detected
      --timeout SECS   How long to wait for each app to quit (default 10)
  -h, --help           Show this help
  -v, --version        Show the version
`

// shells are the shells lightsout has completion scripts for.
var shells = []string{"bash", "zsh", "fish", "powershell"}

// parseArgs reads lightsout's command-line arguments.
//
// A first argument of keep, unkeep or completion selects that command, and
// the rest are its app names or shell. Otherwise arguments that don't start
// with "-" are apps to keep open for this run. --version/-v or --help/-h
// anywhere wins over everything else. Any other argument starting with "-",
// or a missing or invalid value, is an error, so a mistyped option never
// starts a run that quits apps.
func parseArgs(args []string) (options, error) {
	opts := options{cmd: runQuit, timeout: defaultTimeout}
	for _, arg := range args {
		switch arg {
		case "-v", "--version":
			return options{cmd: showVersion}, nil
		case "-h", "--help":
			return options{cmd: showHelp}, nil
		}
	}

	if len(args) > 0 {
		switch args[0] {
		case "keep", "unkeep", "completion":
			return parseSubcommand(args[0], args[1:])
		}
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-n" || arg == "--dry-run":
			opts.dryRun = true
		case arg == "-y" || arg == "--yes":
			opts.yes = true
		case arg == "--timeout":
			if i+1 == len(args) {
				return options{}, fmt.Errorf("--timeout needs a number of seconds")
			}
			i++
			timeout, err := parseTimeout(args[i])
			if err != nil {
				return options{}, err
			}
			opts.timeout = timeout
		case strings.HasPrefix(arg, "--timeout="):
			timeout, err := parseTimeout(strings.TrimPrefix(arg, "--timeout="))
			if err != nil {
				return options{}, err
			}
			opts.timeout = timeout
		case strings.HasPrefix(arg, "-"):
			return options{}, fmt.Errorf("unknown option %q", arg)
		default:
			opts.apps = append(opts.apps, arg)
		}
	}
	return opts, nil
}

// parseSubcommand reads the arguments after keep, unkeep or completion.
func parseSubcommand(name string, args []string) (options, error) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return options{}, fmt.Errorf("unknown option %q for %s", arg, name)
		}
	}
	if len(args) == 0 {
		return options{}, fmt.Errorf("%s needs at least one argument", name)
	}

	switch name {
	case "keep":
		return options{cmd: keepApps, apps: args}, nil
	case "unkeep":
		return options{cmd: unkeepApps, apps: args}, nil
	}
	if len(args) != 1 || !slices.Contains(shells, args[0]) {
		return options{}, fmt.Errorf("completion needs one shell: %s", strings.Join(shells, ", "))
	}
	return options{cmd: printCompletion, apps: args}, nil
}

// parseTimeout reads a --timeout value: whole seconds from 1 to 300.
func parseTimeout(value string) (time.Duration, error) {
	seconds, err := strconv.Atoi(value)
	timeout := time.Duration(seconds) * time.Second
	if err != nil || timeout < time.Second || timeout > maxTimeout {
		return 0, fmt.Errorf("--timeout must be a whole number of seconds from 1 to %d, not %q", int(maxTimeout.Seconds()), value)
	}
	return timeout, nil
}
