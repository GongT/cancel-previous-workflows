# cancel-previous-workflows
A Github action that cancels all previous workflows for older commits in its branch

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
          GITHUB_TOKEN: ${{ github.token }}
```
