# Contributing to lightsout

Thanks for helping improve lightsout. This guide covers reporting problems,
making changes, and getting them merged. Everyone taking part is expected to
follow the [code of conduct](CODE_OF_CONDUCT.md).

## Reporting bugs and requesting features

1. Search the [issues](https://github.com/abhisheksinha1506/lightsout/issues),
   open and closed, to check it hasn't been raised already.
2. Open a new issue with the **Bug report** or **Feature request** template and
   fill in every section.

Good bug reports include your OS and version, how you installed lightsout, the
full terminal output, and your config file if you have one.

## Making a change

1. Comment on the issue you want to work on, or open one first, so nobody
   duplicates your work.
2. Fork the repository. If you have write access, create a branch named after
   the issue number instead.
3. Make your change. Add or update tests for any change in behavior.
4. Run the checks below.
5. Open a pull request that links the issue, and fill in the template. Start
   its title with a [Conventional Commits](https://www.conventionalcommits.org)
   type, such as `feat: add --quiet` or `fix: keep Finder open`: the title
   becomes the commit message, and `feat:` and `fix:` commits start a release
   and appear in its notes.

A maintainer will review it. Branches are deleted automatically after merging.

## Checks

CI runs all of these on every push. Running them first saves a round trip.

| Command            | What it checks                                                    | You need               |
| ------------------ | ----------------------------------------------------------------- | ---------------------- |
| `make build-go`    | lightsout builds                                                  | Go 1.26 or newer       |
| `make test-go`     | the tests pass                                                    | Go 1.26 or newer       |
| `make local-lint`  | golangci-lint, `goreleaser check`, hadolint, and that `go mod tidy` changes nothing | [golangci-lint](https://golangci-lint.run), [GoReleaser](https://goreleaser.com), [hadolint](https://github.com/hadolint/hadolint), git |
| `make docker-lint` | the same lint checks, inside Docker                               | Docker                 |

If you can't run something locally, open the pull request anyway; CI will
show any problems.

Run `make install-hooks` once after cloning, and git will run the tests before
every commit that changes Go code, stopping the commit if they fail.

The tests never quit your apps. Running `lightsout` itself does, so save your
work before trying a change out by hand.

## Writing code

- Format Go code with `gofmt`.
- Keep `go.mod` and `go.sum` tidy. `make lint-gomod` fails if `go mod tidy`
  would change them.
- Tests use the standard `testing` package and must not depend on the network
  or quit real apps. See `pkg/version/version_test.go` for faking HTTP, and
  `main_test.go` for passing in a fake quit function.
- Update the README when you change behavior, flags or configuration.
