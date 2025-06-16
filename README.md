## Robot working on GitHub Actions

:rocket: Goals of the first phase

* [X] `/retest` in PR

* [X] `/[un] assign` in Issue

:memo: Goals of the second phase

* [ ] `/lgtm` in PR 

* [ ] `/[un] cc` in PR

### Quick Start

You can use it in GitHub workflow:

```yaml
name: Issue and PullRequest Command

on:
  issue_comment:
    types:
      - created

jobs:
  actbot:
    runs-on: ubuntu-22.04
    permissions:
      pull-requests: write
      contents: read
      issues: write
    steps:
      - name: Checkout
        uses: actions/checkout@v2
      - uses: ./
        name: Actbot Action
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
```