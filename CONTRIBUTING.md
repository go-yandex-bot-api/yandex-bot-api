# Contributing to yandex-bot-api

First off, thank you for considering contributing to `yandex-bot-api`! It's people like you that make open source such a great community.

## 1. Where do I go from here?

If you've noticed a bug or have a feature request, make one! It's generally best if you get confirmation of your bug or approval for your feature request this way before starting to code.

## 2. Fork & create a branch

If this is something you think you can fix, then fork `yandex-bot-api` and create a branch with a descriptive name.

A good branch name would be (where issue #325 is the ticket you're working on):

```sh
git checkout -b 325-add-new-polling-option
```

## 3. Implementation Guidelines

- **Code Style**: We follow standard Go formatting. Run `gofmt -s -w .` before committing.
- **Linting**: We strictly use `golangci-lint`. Ensure you run it locally and all checks pass.
- **Testing**: Add unit tests for any new features or bug fixes.
- **Documentation**: If you change any exported API, update the GoDoc comments (in English).

To run all checks locally:
```sh
go test -v -race ./...
golangci-lint run ./...
```

## 4. Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/). This means your commit messages should look like this:

- `feat: added new keyboard builder method`
- `fix: resolved data race in router`
- `docs: updated readme examples`
- `test: added tests for webhook package`

## 5. Pull Request

Once you're finished, push your branch and open a Pull Request.
Ensure that the CI pipeline passes (we use GitHub Actions for tests and linting).
