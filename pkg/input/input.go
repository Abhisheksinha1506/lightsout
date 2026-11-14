package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// pickApp prints apps as a numbered list (starting at 1) to out, plus 0 for
// "none of these", and reads the user's choice from in, asking again until a
// listed number is entered.
//
// It returns the chosen app, or "" for 0: the app the user is running from is
// not listed, for example because it is already kept open. It returns an
// error if apps is empty, if out can't be written to, or if in ends or fails
// before a valid number is entered (for example on Ctrl+D).
func pickApp(in io.Reader, out io.Writer, apps []string) (string, error) {
	if len(apps) == 0 {
		return "", errors.New("no running apps to choose from")
	}

	var menu strings.Builder
	menu.WriteString("Which app are you running lightsout from?\n")
	for i, app := range apps {
		menu.WriteString(strconv.Itoa(i+1) + ") " + app + "\n")
	}
	menu.WriteString("0) None of these (the app I'm running from is already kept)\n")
	prompt := "Enter a number (0-" + strconv.Itoa(len(apps)) + "): "

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
		if n, err := strconv.Atoi(answer); err == nil && n >= 0 && n <= len(apps) {
			if n == 0 {
				return "", nil
			}
			return apps[n-1], nil
		}
		message = fmt.Sprintf("%q is not a number from 0 to %d.\n%s", answer, len(apps), prompt)
	}
}

// PickApp asks which of apps lightsout is running from, reading the answer
// from in and writing the question to out. It returns "" if the user chose
// none of them. See pickApp for the errors it returns.
func PickApp(in io.Reader, out io.Writer, apps []string) (string, error) {
	return pickApp(in, out, apps)
}

// Without returns apps minus every name in remove, sorted. Names are compared
// exactly, ignoring surrounding newlines.
func Without(apps, remove []string) []string {
	removed := map[string]bool{}
	for _, name := range remove {
		removed[strings.Trim(name, "\n")] = true
	}
	kept := []string{}
	for _, app := range apps {
		if !removed[strings.Trim(app, "\n")] {
			kept = append(kept, app)
		}
	}
	sort.Strings(kept)
	return kept
}
