# cancel-previous-workflows
Cancels all previous runs for the current workflow, on current branch

## usage
```
name: Cancel

on: push

jobs:
  cancel:
    name: Cancel Previous Runs
    runs-on: ubuntu-latest
    steps:
      - name: cancel running workflows
        uses: GongT/cancel-previous-workflows@master
        env:
          GITHUB_TOKEN: ${{ github.token }}  # Required
		  DELETE: true  # optional, defaults to false, delete all previous runs (including completed one)
```

## tools:

All arguments comes from environment:
```bash
export GITHUB_TOKEN="xxxxxxxxxxxxxxxx"
export GITHUB_REPOSITORY="user/repo"
```

### delete all runs
```bash
go run ./cmd/delete-logs/main.go
```

### delete all runs, except latest
```bash
go run ./cmd/delete-old/main.go
```


### start previously failed runs, but by dispatch_event (will run on latest commit)
```bash
go run ./cmd/delete-old/main.go
```

### start every workflow now, by dispatch_event
```bash
# export FILTER_REGEX="^generated-action-"
go run ./cmd/start-all/main.go
```
