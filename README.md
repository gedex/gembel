gembel
======

ghembel &mdash; command line app to bulk update issue labels GitHub repositories.

## Install

### From binaries

TODO

### From OSX brew

TODO

### From source

TODO

## Using gembel

Before using gembel, you need `GITHUB_TOKEN` (can be retrieved from [here](#)).
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
            "color": "c7def8",
            "replace": "enhancement"
        },
        {
            "name": "[Status] In-Progress"
        }
    ],
    "repositories": [
        "gedex/repo-name",
        "gedex/another-repo-name"
    ]
}
```

It just require `labels` (label properties to apply) and `repositories` (the
target repositories). If label has `replace` property, it will replace matching
label in the repository with the new one in `name`.
