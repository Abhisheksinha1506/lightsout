# Tool images, pinned by tag so Dependabot can keep them up to date.
FROM hadolint/hadolint:v2.15.1 AS hadolint
FROM goreleaser/goreleaser:v2.20.2 AS goreleaser

# golangci-lint image: Go, golangci-lint, git and make.
FROM golangci/golangci-lint:v2.15.0

# Meta data
LABEL maintainer="abhisheksinha1506@gmail.com"
LABEL description="🌙 Quit all applications with one terminal command"

# Installing hadolint and goreleaser
COPY --from=hadolint /bin/hadolint /usr/bin/hadolint
COPY --from=goreleaser /usr/bin/goreleaser /usr/bin/goreleaser

# Copying over files
COPY . /usr/src/app
WORKDIR /usr/src/app

CMD ["make", "local-lint"]
