# AGENTS.md

## Build & Dev Commands

```bash
make build        # fmt + vet + lint then compile to ./bin/lazyssh
make run          # go run ./cmd/main.go (no quality gates)
make test         # go test -race -coverprofile=coverage.out ./...
make lint         # gofumpt + golangci-lint (installs tools to ./bin/ automatically)
make fmt          # gofumpt -l -w . && go fmt ./...
make quality      # fmt -> vet -> lint (run this before pushing)
make check        # staticcheck ./...
```

`make build` runs the full `quality` chain first. For a quick compile-only check, use `go build ./cmd` directly.

Tools (`golangci-lint`, `gofumpt`, `staticcheck`) are installed into `./bin/` on first use by the Makefile; no global install required.

## License Header (enforced by linter)

Every `.go` file must start with this exact Apache 2.0 header (enforced by `goheader` in `.golangci.yml`):

```go
// Copyright 2025.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
```

## Architecture

Hexagonal / ports-and-adapters layout:

```
cmd/main.go                          # entrypoint, wires everything
internal/
  core/
    domain/server.go                 # Server model
    ports/                           # interfaces (ServerRepository, ServerService)
    services/server_service.go       # business logic + SSH exec (platform-specific sysprocattr_*.go)
  adapters/
    data/ssh_config_file/            # reads/writes ~/.ssh/config and ~/.lazyssh/metadata.json
    ui/                              # tview TUI (server list, form, search, handlers)
  logger/                            # zap wrapper
```

- `ports/` defines interfaces; `services/` implements `ServerService`; `data/` implements `ServerRepository`.
- The TUI layer is in `adapters/ui/` and uses `tview`/`tcell`.
- The SSH config parser is a **forked dependency**: `github.com/kevinburke/ssh_config` is replaced by `github.com/adembc/ssh_config v1.4.2` (see `go.mod` replace directive).

## Formatter

Code is formatted with **gofumpt** (stricter than `gofmt`). Run `make fmt` or the linter will fail.

## PR Title Convention (CI-enforced)

Format: `type(scope): description`

- **Types:** `feat`, `fix`, `improve`, `refactor`, `docs`, `test`, `ci`, `chore`, `revert`
- **Scopes (optional):** `ui`, `cli`, `config`, `parser`

## CI

On PR/push to `main`: `make build` then `make test`. Semantic PR title check runs separately.
