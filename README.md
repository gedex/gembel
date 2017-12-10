gembel
======

gembel &mdash; command line app to bulk update GitHub issue labels.

## Install

### From brew

```
brew install gedex/tap/gembel
```

Check the [tap source](https://github.com/gedex/homebrew-tap) for more details.

### From binaries

Download your preferred flavor from the [releases page](https://github.com/gedex/gembel/releases/latest) and install manually.

### From Go Get

```
go get github.com/gedex/gembel
```

## Using gembel

Before using gembel, you need `GITHUB_TOKEN` (can be retrieved from [here](https://github.com/settings/tokens)).
Once you've that, set it to your bash profile or provide it when running the app:

```
GITHUB_TOKEN="token" gembel <config-file>
```

`<config-file>` is JSON file with following structure:


```json
{
    "labels": [
        {
            "name": "[Type] Bug",
            "color": "e11d21",
            "replace": "bug"
        },
        {
            "name": "[Type] Enhancement",
            "color": "c7def8"
        }
    ],
    "repositories": [
        "gedex/repo-name",
        "gedex/another-repo-name"
    ]
}
```

It requires `labels` (label properties to apply) and `repositories` (the
target repositories). If label has `replace` property (optional), it will replace
matching label in the repository with the new one in `name`.
