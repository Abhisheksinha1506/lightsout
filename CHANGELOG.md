# Changelog

All notable changes to lightsout are listed here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- lightsout detects the app you're running it from, by following its own
  parent processes, and keeps it open without asking. It still asks when it
  can't tell, for example over SSH or inside tmux.
- `--dry-run` (`-n`) lists the apps that would be asked to quit, and quits
  nothing.
- `--yes` (`-y`) never asks which app you're running from. If that app can't
  be detected, lightsout stops without quitting anything, so it is safe to
  run from scripts and schedules.
- `--timeout SECS` sets how long to wait for each app to quit (1–300 seconds,
  default 10).
- `lightsout keep APP...` and `lightsout unkeep APP...` add and remove apps in
  `ignoredApps`, creating the config file if needed.
- Shell completions for bash, zsh, fish and PowerShell
  (`lightsout completion SHELL`) and a man page, installed by Homebrew and
  the `.deb`/`.rpm` packages.
- Releases include SBOMs, a cosign signature for the checksums file, and
  GitHub build provenance for every file.
- `SECURITY.md`, issue forms, a demo GIF and a FAQ in the README.

### Changed

- Error messages at the end of a run start in lowercase, for example
  "failed to quit: Slack".
- CI tests the oldest supported Go as well as the newest, runs an end-to-end
  test on macOS too, installs the `.deb` and `.rpm` packages, and runs
  govulncheck, CodeQL and OpenSSF Scorecard. Every GitHub Action is pinned to
  a commit.

## [1.1.4] - 2026-11-14

### Added

- `--version` (`-v`) prints the version and `--help` (`-h`) prints usage.

### Fixed

- An argument starting with `-` no longer starts a run. Before,
  `lightsout --version` treated `--version` as an app name to keep open and
  went on to ask every other app to quit. Unknown options now print an error
  and exit with status 2.

## [1.1.3] - 2026-11-14

No changes to how lightsout itself behaves.

### Added

- Each release opens a pull request to `microsoft/winget-pkgs` with the new
  version's manifest, so Windows users can install and upgrade with winget.
- `.deb` and `.rpm` packages for Linux on each release. They install the
  `wmctrl` and `xprop` tools lightsout needs.
- A Scoop bucket for Windows, updated by each release:
  `scoop bucket add abhisheksinha1506 https://github.com/abhisheksinha1506/scoop-bucket`.
- The README explains installing with `go install`.

## [1.1.2] - 2026-11-14

### Fixed

- The update check compares version numbers instead of text, so it only asks
  you to update when a newer release exists. Before, any difference (including
  running a newer or prerelease build) showed "PLEASE UPDATE".

## [1.1.1] - 2026-11-14

### Changed

- The repository is public, so `brew install abhisheksinha1506/tap/lightsout`
  and release downloads no longer need a GitHub token, and the update check
  can see new releases.

## [1.1.0] - 2026-11-14

### Added

- Windows support. lightsout lists the apps on your taskbar, asks each to close
  the same way as clicking its X button, and closes File Explorer windows at
  the end. Releases include `.zip` archives for Windows on amd64 and arm64.
- GitHub release notes are now taken from this changelog, and a release stops
  before publishing anything if its version has no section here.

## [1.0.3] - 2026-11-14

No changes to how lightsout itself behaves.

### Added

- A pre-commit hook that runs `go test ./...` before any commit that changes
  Go code. Turn it on once per clone with `make install-hooks`.
- This changelog, which release archives now include.

### Changed

- The README explains that release downloads and the update check need access
  to the repository while it is private, and documents `make install-hooks`.

## [1.0.2] - 2026-11-14

### Added

- Choice `0` ("None of these") in the "Which app are you running lightsout
  from?" prompt, for when that app is already in `ignoredApps`. Before, you
  had to pick another app, which was then kept open by mistake.

### Fixed

- When every running app is already kept open, lightsout no longer asks which
  app it is running from. It used to exit with "Failed to get running
  terminal".

## [1.0.1] - 2026-11-14

### Added

- Homebrew installs from the private repository. The cask downloads release
  files through the GitHub API with `HOMEBREW_GITHUB_API_TOKEN`, sent only as a
  request header.

## [1.0.0] - 2026-11-14

First release.

### Added

- The `lightsout` command, which asks every open app to quit except the
  terminal you run it from and the apps you choose to keep.
- macOS: lists the apps in your Dock, sends each the normal Quit request by
  process ID, and closes Finder windows.
- Linux: closes each app's windows with `wmctrl`, skipping desktop and panel
  windows, and warns in Wayland sessions.
- A numbered prompt for the app you are running lightsout from, with kitty
  detected automatically.
- Apps still open after 10 seconds are left running and listed at the end, and
  lightsout exits with status 1.
- A config file, `~/.config/lightsout/config.yml`, with `ignoredApps` and
  `ignoredUpdates`. App names can also be passed as arguments.
- An update check for release builds that gives up after 5 seconds.
- Releases for macOS and Linux built with GoReleaser, a Homebrew cask, CI, and
  Dependabot.

[Unreleased]: https://github.com/abhisheksinha1506/lightsout/compare/v1.1.4...HEAD
[1.1.4]: https://github.com/abhisheksinha1506/lightsout/compare/v1.1.3...v1.1.4
[1.1.3]: https://github.com/abhisheksinha1506/lightsout/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/abhisheksinha1506/lightsout/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/abhisheksinha1506/lightsout/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.3...v1.1.0
[1.0.3]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/abhisheksinha1506/lightsout/releases/tag/v1.0.0
