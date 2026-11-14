# lightsout

[![Latest release](https://img.shields.io/github/v/release/abhisheksinha1506/lightsout)](https://github.com/abhisheksinha1506/lightsout/releases/latest)
[![build](https://github.com/abhisheksinha1506/lightsout/actions/workflows/build.yml/badge.svg)](https://github.com/abhisheksinha1506/lightsout/actions/workflows/build.yml)
[![lint](https://github.com/abhisheksinha1506/lightsout/actions/workflows/lint.yml/badge.svg)](https://github.com/abhisheksinha1506/lightsout/actions/workflows/lint.yml)
[![release](https://github.com/abhisheksinha1506/lightsout/actions/workflows/release.yml/badge.svg)](https://github.com/abhisheksinha1506/lightsout/actions/workflows/release.yml)

🌙 Quit all applications with one terminal command.

lightsout asks every open desktop application on macOS or Linux to quit, except
the terminal you run it from and any apps you tell it to keep.

> [!NOTE]
> lightsout asks apps to quit the normal way, so an app with unsaved work can
> ask you to save first. Any app still open after 10 seconds is left running
> and listed at the end. Nothing is force quit.

## Install

### macOS

```bash
brew install abhisheksinha1506/tap/lightsout
```

While the repository is private, Homebrew needs a GitHub token that can read
it: export `HOMEBREW_GITHUB_API_TOKEN` in your shell before installing or
upgrading.

The first time you run lightsout, macOS asks whether your terminal may control
Finder, so it can close Finder's windows. Allow it; macOS only asks once.

### Linux

Download the archive for your architecture (amd64, arm64, armv6 or armv7) from
the [releases page](https://github.com/abhisheksinha1506/lightsout/releases)
(you need access to the repository while it is private), extract it, and put
`lightsout` somewhere on your `PATH`.

lightsout finds windows with `wmctrl` and `xprop`, so install both. On Debian
or Ubuntu:

```bash
sudo apt install wmctrl x11-utils
```

These tools work on X11 desktops. In a Wayland session, only apps running
through XWayland can be seen and closed; lightsout warns you when it detects
one.

## Usage

```bash
lightsout
```

1. lightsout lists the running applications, numbered: on macOS, the apps in
   your Dock; on Linux, the apps with open windows.
2. It asks which app you are running it from, so it can keep that one open.
   Type that app's number and press <kbd>Enter</kbd>. If that app isn't
   listed because it's already in `ignoredApps`, type <kbd>0</kbd> (none of
   these). If you use [kitty](https://sw.kovidgoyal.net/kitty/), lightsout
   detects it and skips the question; if every app is already kept, there is
   nothing to ask.
3. It asks every other app to quit and waits up to 10 seconds for each. On
   macOS it also closes all Finder windows.

Press <kbd>Ctrl</kbd>+<kbd>C</kbd> at the question to cancel before anything
is quit.

To keep more apps open, pass their names as arguments. Use the names exactly as
lightsout lists them, and quote names that contain spaces:

```bash
lightsout Slack "Google Chrome"
```

If an app can't be quit, or is still open after 10 seconds (for example because
it is asking you to save), lightsout prints a warning and carries on with the
rest. At the end it lists every app it couldn't quit and exits with status 1.

## Configuration

lightsout reads an optional config file from `~/.config/lightsout/config.yml`
(or `config.yaml`). Create the file yourself if you need it. If your config
file lives in another folder from an older release, move it here, or its
settings won't apply.

```yaml
# Apps lightsout never quits
ignoredApps:
  - Google Chrome
  - Terminal

# Skip the update check on every run
ignoredUpdates: true
```

| Key              | Type            | Default | Effect                                                     |
| ---------------- | --------------- | ------- | ---------------------------------------------------------- |
| `ignoredApps`    | list of strings | none    | Apps to keep open. Command-line arguments are added to it. |
| `ignoredUpdates` | boolean         | `false` | Turns off the update check.                                |

If the config file exists but can't be read, lightsout stops before quitting
anything.

## Update check

Release builds ask GitHub for the latest version on each run and print a notice
when a newer one is out. If GitHub can't be reached, takes longer than 5
seconds to answer, or the check fails, lightsout carries on without it (with a
warning when GitHub answered but the check failed). Builds made from source
skip the check.

While the repository is private, the check can't see its releases and warns on
every run, so set `ignoredUpdates: true` in your config to turn it off.

## How it works

| Platform | Lists apps with                                                                  | Asks each app to quit with                         |
| -------- | -------------------------------------------------------------------------------- | -------------------------------------------------- |
| macOS    | Dock apps from `NSWorkspace` (JavaScript for Automation through `osascript`)     | The normal Quit request, sent to each process ID   |
| Linux    | Windows from `wmctrl -l`, grouped by `WM_CLASS`; desktop and panel windows skipped | A close request to each of its windows (`wmctrl -i -c`) |

lightsout then waits for the app to exit (macOS) or for its windows to close
(Linux). On macOS, Finder windows are then closed with AppleScript; Finder
itself keeps running.

## Development

You need Go 1.26 or newer. CI uses the latest stable release, and Dependabot
opens a weekly pull request for each of Go modules, GitHub Actions and the
lint Docker images.

```bash
make build-go       # download modules and check that lightsout builds
make test-go        # run the tests
make install-hooks  # run the tests before every commit that changes Go code
make local-lint     # golangci-lint, goreleaser check, hadolint and a go.mod tidy check
make docker-lint    # the same lint checks inside Docker, with nothing else to install
```

Run `make install-hooks` once after cloning; a commit is stopped if the tests
fail.

`make local-lint` needs [golangci-lint](https://golangci-lint.run),
[GoReleaser](https://goreleaser.com) and
[hadolint](https://github.com/hadolint/hadolint) installed.

The tests never quit any apps. Running `lightsout` does, so only run it when
you are ready for everything to close.

## Releasing

First move the notes under "Unreleased" in [CHANGELOG.md](CHANGELOG.md) into a
new section for the version, and commit that. Then push a tag such as
`v1.1.0`. The release workflow first turns that CHANGELOG.md section into the
release notes (`go run ./tools/releasenotes <tag>`), and stops before
publishing anything if the section is missing. Then it runs GoReleaser, which:

- builds the binaries and publishes them as a GitHub release, with those notes,
- stamps the tag into each binary as its version, which the update check uses,
  and
- updates the Homebrew cask in
  [abhisheksinha1506/homebrew-tap](https://github.com/abhisheksinha1506/homebrew-tap).

The workflow needs a `GORELEASER_PAT` secret that can write to both
repositories.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for what changed in each release.

## Contributing

Bug reports, ideas and pull requests are welcome. Read
[CONTRIBUTING.md](CONTRIBUTING.md) first, and follow the
[code of conduct](CODE_OF_CONDUCT.md).

## License

lightsout is released under the [MIT License](LICENSE.md).

