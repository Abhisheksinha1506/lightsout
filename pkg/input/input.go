package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gleich/statuser/v2"
)

// pickApp prints apps as a numbered list (starting at 1) to out and reads the
// user's choice from in, asking again until a listed number is entered.
//
// It returns the chosen app. It returns an error if apps is empty, if out
// can't be written to, or if in ends or fails before a valid number is entered
// (for example on Ctrl+D).
func pickApp(in io.Reader, out io.Writer, apps []string) (string, error) {
	if len(apps) == 0 {
		return "", errors.New("no running apps to choose from")
	}

	var menu strings.Builder
	menu.WriteString("Which app are you running lightsout from?\n")
	for i, app := range apps {
		menu.WriteString(strconv.Itoa(i+1) + ") " + app + "\n")
	}
	prompt := "Enter a number (1-" + strconv.Itoa(len(apps)) + "): "

	message := menu.String() + prompt
	scanner := bufio.NewScanner(in)
	for {
		if _, err := io.WriteString(out, message); err != nil {
			return "", err
		}
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", errors.New("no app chosen")
		}
		answer := strings.TrimSpace(scanner.Text())
		if n, err := strconv.Atoi(answer); err == nil && n >= 1 && n <= len(apps) {
			return apps[n-1], nil
		}
		message = fmt.Sprintf("%q is not a number from 1 to %d.\n%s", answer, len(apps), prompt)
	}
}

// Ask the user what terminal they are executing from
func ExecutingTerm(runningApps, ignoredApps []string) []string {
	cleanedApps := []string{}

	// Removing ignored applications
	for _, app := range runningApps {
		var isIgnored bool
		for _, ignoredApp := range ignoredApps {
			if strings.Trim(app, "\n") == strings.Trim(ignoredApp, "\n") {
				isIgnored = true
			}
		}
		if !isIgnored {
			cleanedApps = append(cleanedApps, app)
		}
	}
	sort.Strings(cleanedApps)

	// Automatically select terminal as kitty if $TERM is xterm-kitty
	var program string
	for _, app := range runningApps {
		if app == "kitty" && os.Getenv("TERM") == "xterm-kitty" {
			statuser.Success("Kitty terminal automatically detected")
			program = "kitty"
		}
	}

	if program == "" {
		var err error
		program, err = pickApp(os.Stdin, os.Stdout, cleanedApps)
		if err != nil {
			statuser.Error("Failed to get running terminal", err, 1)
		}
	}

	// program is always running: it was picked from cleanedApps or detected in
	// runningApps. It may already be missing here if it is also ignored.
	cleanedApps2 := []string{}
	for _, app := range cleanedApps {
		if strings.Trim(app, "\n") != strings.Trim(program, "\n") {
			cleanedApps2 = append(cleanedApps2, app)
		}
	}
	return cleanedApps2
}
