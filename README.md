# cete

Cete is a CLI to pit UCI-compliant chess engines against each other.

## Installation

Installation is only done via the `go` command for now:

```sh
go install github.com/leonhfr/cete@latest
```

## Quick start

Play a game using flags:

```sh
# Engines can be binaries in the PATH or file paths to the binaries:
cete --white stockfish --black ./honeybadger
```

Play a game using a configuration file:

```sh
cete game ./test/data/stockfish.yaml
```

An example of a configuration can be found in `/test/data`.
