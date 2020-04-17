# Contribution guide

Contributions to Gnomock and its ecosystem can be of any kind, but most of them
will probably be one of these:

1. [Bug reports](#bug-reports)
2. [Feature requests](#feature-requests)
3. [Pull requests](#pull-requests)
4. [Code review](#code-review)

## Bug reports

Gnomock like any other software has bugs, or behaves differently from what you
expect due to missing documentation. To report such cases, please open an issue
either in Gnomock main repository, or in the repository of an involved preset.

Please make sure to include steps to reproduce the issue you are facing. The
best way would be to link a repository that can be cloned. Any sufficient
code snippet inside the issue itself would also be fine.

## Feature requests

Gnomock should always move forward and add features that could make its users'
lives easier, and tests – better. If you think there is something that can help
the ecosystem, please submit a new issue either in Gnomock main repository, or
in the repository of an involved preset.

In the issue, please describe what would you like to see implemented, and why.
The best way would be by example: just write some code that uses the new
feature as if it existed, and describe what happens when the code runs.

If such an example is impossible to write, please try to explain your needs,
and we would try to come up with something together.

## Pull requests

If you decide to contribute actual code to one of Gnomock's repositories,
please let others know about it by opening a feature request you'd like to
implement, or a bug report you'd like to fix. This issue can be used to discuss
possible solutions. If you end up writing actual code, please make sure that
you follow existing conventions:

1. [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

2. Run [`golangci-lint`](https://github.com/golangci/golangci-lint) on your
   code using [.golangci.yml](.golangci.yml) file

3. Make sure all existing tests/examples pass, and add/modify tests to cover
   your changes

4. Document your changes: Go doc, example code in [README.md](README.md), pull
   request description, linked issue, etc.

5. Avoid breaking changes

## Code review

You are welcome to review existing pull requests, or existing code. Please be
civil, and try to back up your comments with whatever may help to convince the
other parties involved. For example, when pointing out a bug, present a way to
reproduce it (even if it is theoretical).
