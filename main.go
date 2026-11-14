package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/config"
	"github.com/abhisheksinha1506/lightsout/pkg/input"
	"github.com/abhisheksinha1506/lightsout/pkg/output"
	"github.com/abhisheksinha1506/lightsout/pkg/version"
	"github.com/briandowns/spinner"
	"github.com/gleich/statuser/v2"
)

//go:embed completions/*
var completions embed.FS

// completionFiles maps each shell in shells to its script in completions,
// named the way that shell expects to find it.
var completionFiles = map[string]string{
	"bash":       "completions/lightsout.bash",
	"zsh":        "completions/_lightsout",
	"fish":       "completions/lightsout.fish",
	"powershell": "completions/lightsout.ps1",
}

// app is one running app and the process IDs it runs as.
type app struct {
	name string
	pids []int
}

// platform is what run needs from the operating system. Each supported OS
// has a platform_<os>.go file whose newPlatform fills it in; tests use fakes.
type platform struct {
	list          func() ([]app, error)                          // the running apps
	quit          func(name string, timeout time.Duration) error // ask one listed app to quit
	afterQuit     func()                                         // cleanup once every app was asked to quit
	afterQuitNote string                                         // what afterQuit does, for --dry-run
	warning       string                                         // apps lightsout can't see, or ""
	ancestors     []int                                          // lightsout's process and its ancestors
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "lightsout:", err)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	statuser.Emojis = false

	switch opts.cmd {
	case showVersion:
		fmt.Println("lightsout", version.Version)
		return
	case showHelp:
		fmt.Print(usage)
		return
	case printCompletion:
		script, _ := completions.ReadFile(completionFiles[opts.apps[0]])
		os.Stdout.Write(script) //nolint:errcheck // nothing to do if stdout is gone
		return
	case keepApps, unkeepApps:
		if err := editKept(opts); err != nil {
			statuser.Error("Failed to update the config", err, 1)
		}
		return
	}

	if !platformSupported {
		statuser.ErrorMsg("This application only works on macOS, Linux and Windows", 1)
	}

	var conf config.Conf
	if config.Exists() {
		config.Read(&conf)
	}
	keep := append(conf.IgnoredApps, opts.apps...)

	output.Title()
	if !conf.IgnoreUpdates {
		version.CheckForUpdate()
	}

	if err := run(opts, keep, newPlatform(), os.Stdin, os.Stdout); err != nil {
		statuser.ErrorMsg(err.Error(), 1)
	}
}

// editKept adds (keep) or removes (unkeep) opts.apps in the config's
// ignoredApps and says what changed.
func editKept(opts options) error {
	if opts.cmd == keepApps {
		added, err := config.AddIgnoredApps(opts.apps)
		if err != nil {
			return err
		}
		if len(added) == 0 {
			fmt.Println("Already kept open:", strings.Join(opts.apps, ", "))
			return nil
		}
		statuser.Success("Always keeping open: " + strings.Join(added, ", "))
		return nil
	}

	removed, err := config.RemoveIgnoredApps(opts.apps)
	if err != nil {
		return err
	}
	if len(removed) == 0 {
		fmt.Println("Not in the kept apps:", strings.Join(opts.apps, ", "))
		return nil
	}
	statuser.Success("No longer always keeping open: " + strings.Join(removed, ", "))
	return nil
}

// run lists the running apps, works out which one lightsout is running from
// (asking on in/out if it can't tell), and asks every other app not in keep
// to quit. With opts.dryRun it only prints what it would quit.
//
// It returns an error, after trying every app, if any app is still open; if
// the apps can't be listed; or if it can't tell which app it is running from
// and can't ask (opts.yes, or the question was cancelled).
func run(opts options, keep []string, p platform, in io.Reader, out io.Writer) error {
	s := spinner.New(spinner.CharSets[13], 30*time.Millisecond)
	s.Suffix = " Getting list of applications"
	s.Start()
	apps, err := p.list()
	s.Stop()
	if err != nil {
		return fmt.Errorf("failed to get the running applications: %w", err)
	}
	if p.warning != "" {
		statuser.Warning(p.warning)
	}

	names := make([]string, 0, len(apps))
	for _, a := range apps {
		names = append(names, a.name)
	}
	sort.Strings(names)
	candidates := input.Without(names, keep)

	terminal := runningFrom(apps, p.ancestors)
	if terminal == "" && os.Getenv("TERM") == "xterm-kitty" && slices.Contains(names, "kitty") {
		terminal = "kitty"
	}

	var toQuit []string
	switch {
	case terminal != "":
		statuser.Success("Keeping " + terminal + " open (you're running lightsout from it)")
		toQuit = input.Without(candidates, []string{terminal})
	case len(candidates) == 0:
		toQuit = candidates
	case opts.dryRun:
		say(out, "lightsout couldn't tell which app you're running it from, so it would ask you.")
		toQuit = candidates
	case opts.yes:
		return errors.New("couldn't tell which app you're running lightsout from, so nothing was quit. " +
			"Run it without --yes to choose, or keep that app with: lightsout keep <app>")
	default:
		picked, err := input.PickApp(in, out, candidates)
		if err != nil {
			return fmt.Errorf("failed to get running terminal: %w", err)
		}
		toQuit = input.Without(candidates, []string{picked})
	}
	say(out)

	if opts.dryRun {
		printDryRun(out, toQuit, p.afterQuitNote)
		return nil
	}

	failed := quitApps(toQuit, func(name string) error { return p.quit(name, opts.timeout) })
	p.afterQuit()

	say(out)
	if len(failed) > 0 {
		return errors.New("failed to quit: " + strings.Join(failed, ", "))
	}
	statuser.Success("Quit All Applications! Have a good day :)")
	return nil
}

// say writes a line to out. Errors are ignored: out is the terminal, and if
// it is gone there is nobody left to tell.
func say(out io.Writer, a ...any) { _, _ = fmt.Fprintln(out, a...) }

// printDryRun lists the apps a real run would ask to quit.
func printDryRun(out io.Writer, toQuit []string, afterQuitNote string) {
	if len(toQuit) == 0 {
		say(out, "Dry run: no apps would be quit.")
	} else {
		say(out, "Dry run: these apps would be asked to quit:")
		for _, name := range toQuit {
			say(out, "  •", name)
		}
	}
	if afterQuitNote != "" {
		say(out, "It would also", afterQuitNote+".")
	}
	say(out, "Nothing was quit.")
}

// runningFrom returns the app lightsout is running from: the listed app that
// owns the closest of ancestors (lightsout's own process first, then its
// parent and so on), or "" if none of them belongs to a listed app.
func runningFrom(apps []app, ancestors []int) string {
	owner := map[int]string{}
	for _, a := range apps {
		for _, pid := range a.pids {
			owner[pid] = a.name
		}
	}
	for _, pid := range ancestors {
		if name, ok := owner[pid]; ok {
			return name
		}
	}
	return ""
}

// quitApps quits each app in order using quit, showing a spinner per app.
// A failure is printed as a warning and skipped so the remaining apps are
// still quit. It returns the apps that failed to quit, in order.
func quitApps(apps []string, quit func(app string) error) []string {
	var failed []string
	for _, app := range apps {
		s := spinner.New(spinner.CharSets[13], 30*time.Millisecond)
		s.Suffix = " Quitting " + app
		s.Start()
		err := quit(app)
		s.Stop()
		if err != nil {
			statuser.Warning("Failed to quit "+app+":", err)
			failed = append(failed, app)
			continue
		}
		statuser.Success("Quit " + app)
	}
	return failed
}
