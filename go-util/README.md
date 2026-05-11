# go-util

Some useful (or useless) scripts written in Go.

## Scripts index

| Script | Description | Key env vars |
|--------|-------------|-------------|
| `gitea-delete-orgs` | Delete all Gitea organisations except a keep-list | `GITEA_URL`, `GITEA_TOKEN`, `GITEA_KEEP_ORGS` |

## Structure

Each script lives in its own directory under `cmd/`, so multiple `main` packages coexist without conflicts.

```
cmd/
  gitea-delete-orgs/   # Delete all Gitea orgs except a keep-list
```

## Usage

```bash
# Run any script
go run ./cmd/<script-name>

# Example: delete Gitea orgs
export GITEA_URL=https://gitea.example.com
export GITEA_TOKEN=your-token
export GITEA_KEEP_ORGS=org1,org2
go run ./cmd/gitea-delete-orgs
```

## Adding a new script

Create a new directory under `cmd/` with its own `main.go`:

```bash
mkdir cmd/my-new-script
# write cmd/my-new-script/main.go with `package main`
go run ./cmd/my-new-script
```
