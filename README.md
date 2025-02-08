<img align=right width="168" src="docs/gouef_logo.png">

# githubtoplanguages
Generate your Top languages of Github

[![Static Badge](https://img.shields.io/badge/Github-gouef%2Fgithubtoplanguages-blue?style=for-the-badge&logo=github&link=github.com%2Fgouef%2Fgithubtoplanguages)](https://github.com/gouef/githubtoplanguages)

[![GoDoc](https://pkg.go.dev/badge/github.com/gouef/githubtoplanguages.svg)](https://pkg.go.dev/github.com/gouef/githubtoplanguages)
[![GitHub stars](https://img.shields.io/github/stars/gouef/githubtoplanguages?style=social)](https://github.com/gouef/githubtoplanguages/stargazers)
[![Go Report Card](https://goreportcard.com/badge/github.com/gouef/githubtoplanguages)](https://goreportcard.com/report/github.com/gouef/githubtoplanguages)
[![codecov](https://codecov.io/github/gouef/githubtoplanguages/branch/main/graph/badge.svg?token=YUG8EMH6Q8)](https://codecov.io/github/gouef/githubtoplanguages)

## Versions
![Stable Version](https://img.shields.io/github/v/release/gouef/githubtoplanguages?label=Stable&labelColor=green)
![GitHub Release](https://img.shields.io/github/v/release/gouef/githubtoplanguages?label=RC&include_prereleases&filter=*rc*&logoSize=diago)
![GitHub Release](https://img.shields.io/github/v/release/gouef/githubtoplanguages?label=Beta&include_prereleases&filter=*beta*&logoSize=diago)


## Example

![Github Top Languages](https://raw.githubusercontent.com/gouef/githubtoplanguages/refs/heads/main/toplanguages.svg)

```markdown
![Github Top Languages](https://raw.githubusercontent.com/gouef/githubtoplanguages/refs/heads/main/toplanguages.svg)
```

### Action

```yaml
name: Create Top Languages

on:
  push:
  schedule:
    - cron: '0 */2 * * *'

jobs:
  build-and-run:
    runs-on: ubuntu-latest
    steps:
      - name: Run custom action
        uses: gouef/githubtoplanguages@main
        with:
          botName: "Jan Galek"
          botEmail: "ghome.cz@gmail.com"
          user: "JanGalek"
          limit: 6
          ignoredOrgsFlag: "wowmua"
          ignoredReposFlag: "wowmua/Maps"
        env:
          GITHUB_TOKEN: ${{ secrets.USER_GITHUB_TOKEN }}
```

## Action inputs

| Input              | Default              | Required | Description                                   |
|--------------------|----------------------|----------|-----------------------------------------------|
| `botName`          | `Jan Galek`          | `false`  | Set bot name for contributors                 |
| `botEmail`         | `ghome.cz@gmail.com` | `false`  | Set bot email address for contributors        |
| `user`             | ``                   | `true`   | Github UserName                               |
| `limit`            | `6`                  | `true`   | Limit of languages                            |
| `ignoredOrgsFlag`  | ``                   | `true`   | Comma-separated list of ignored organizations |
| `ignoredReposFlag` | ``                   | `true`   | Comma-separated list of ignored repositories  |


## Contributing

Read [Contributing](CONTRIBUTING.md)

## Contributors

<div>
<span>
  <a href="https://github.com/JanGalek"><img src="https://raw.githubusercontent.com/gouef/githubtoplanguages/refs/heads/contributors-svg/.github/contributors/JanGalek.svg" alt="JanGalek" /></a>
</span>
<span>
  <a href="https://github.com/actions-user"><img src="https://raw.githubusercontent.com/gouef/githubtoplanguages/refs/heads/contributors-svg/.github/contributors/actions-user.svg" alt="actions-user" /></a>
</span>
</div>

## Join our Discord Community! 🎉

[![Discord](https://img.shields.io/discord/1334331501462163509?style=for-the-badge&logo=discord&logoColor=white&logoSize=auto&label=Community%20discord&labelColor=blue&link=https%3A%2F%2Fdiscord.gg%2FwjGqeWFnqK
)](https://discord.gg/wjGqeWFnqK)

Click above to join our community on Discord!
