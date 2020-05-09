# <div align="center">Gnomock daemon - use Gnomock with any language</div>

`gnomockd` is a simple HTTP wrapper for
[Gnomock](https://github.com/orlangure/gnomock). A running daemon allows any
code written in any language to interact with Gnomock and its Presets using
HTTP calls.

## <div align="center">![Build](https://github.com/orlangure/gnomockd/workflows/Build/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/orlangure/gnomockd)](https://goreportcard.com/report/github.com/orlangure/gnomockd)</div>

## Dev dependencies

`gnomockd` does not have any runtime dependencies. For development, there are
some requirements though, and this list will be growing as more SDKs are
implemented:

1. Python 3 with `venv` support is required to test Python SDK alongside
   regular `gnomockd` tests.
