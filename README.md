## tailscale-helper

`tailscale-helper` is a small Go CLI that runs `tailscale status --json`, normalizes the result, and prints a more readable table or JSON output.

## Features

- Shows Tailscale status in a compact table
- Supports JSON output for scripting
- Preserves useful peer details such as hostname, owner, OS, activity, and exit-node state

## Requirements

- Go 1.26+
- `tailscale` installed and available in `PATH`
- An authenticated local Tailscale client

## Run

```bash
go run .
```

Show JSON output:

```bash
go run . --json
```

Or use the explicit command:

```bash
go run . status
```

## Build

```bash
go build -o bin/tailscale-helper .
```

Install using go install
```bash
go install github.com/romaingallez/tailscale-helper@latest
```

## Disclaimer

Some parts of this codebase were written or refined with the help of AI agents. All generated code should be reviewed and evaluated with the same care as any other contribution.
