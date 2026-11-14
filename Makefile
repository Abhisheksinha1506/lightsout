##########
# Building
##########

build-docker-dev-lint:
	docker build -f dev.lint.Dockerfile .
build-go:
	go mod download
	go build -v .
	rm lightsout

#########
# Testing
#########

test-go:
	go test ./...
# Run the tests before every commit (see .githooks/pre-commit)
install-hooks:
	git config core.hooksPath .githooks

#########
# Linting
#########

lint-golangci:
	golangci-lint run
lint-gomod:
	go mod tidy
	git diff --exit-code go.mod
	git diff --exit-code go.sum
lint-goreleaser:
	goreleaser check
lint-hadolint:
	hadolint dev.lint.Dockerfile
lint-in-docker:
	docker build -f dev.lint.Dockerfile -t lightsout:lint .
	docker run lightsout:lint

##########
# Grouping
##########

# Linting
local-lint: lint-golangci lint-goreleaser lint-hadolint lint-gomod
docker-lint: lint-in-docker
# Build
local-build: build-docker-dev-lint
