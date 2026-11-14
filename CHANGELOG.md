# Changelog

All notable changes to lightsout are listed here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

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

[Unreleased]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.3...HEAD
[1.0.3]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/abhisheksinha1506/lightsout/releases/tag/v1.0.0
