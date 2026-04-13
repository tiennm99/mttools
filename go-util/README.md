# go-util

Some useful (or useless) scripts written in Go.

## Structure

Each script lives in its own directory under `cmd/`, so multiple `main` packages coexist without conflicts.

```
cmd/
  gitea-leave-orgs/   # Leave all Gitea orgs except a keep-list
```

## Usage

```bash
# Run any script
go run ./cmd/<script-name>

# Example: leave Gitea orgs
export GITEA_URL=https://gitea.example.com
export GITEA_TOKEN=your-token
export GITEA_KEEP_ORGS=org1,org2
go run ./cmd/gitea-leave-orgs
```

## Adding a new script

Create a new directory under `cmd/` with its own `main.go`:

```bash
mkdir cmd/my-new-script
# write cmd/my-new-script/main.go with `package main`
go run ./cmd/my-new-script
```
