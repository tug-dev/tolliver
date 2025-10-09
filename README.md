# Tolliver

> [!WARNING]
> This library is work in progress and anything can change at any time.

A message passing Rust library for sending both fast messages and those that require deliverability guarantees.

## Usage

Add Tolliver to you `Cargo.toml` like this:
```toml
tolliver = { git = "https://github.com/tug-dev/tolliver" }
```

And add the SQLite file to your `.gitignore`
```
tolliver.db
```

## CLI

For more information on the interactive CLI see [docs/cli.md](docs/cli.md).

## Protocol

To see the specification for the Tolliver protocol, see [docs/protocol.md](docs/protocol.md).
