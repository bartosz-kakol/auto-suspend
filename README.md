# auto-suspend

A daemon that automatically suspends your computer based on the output of user-defined Python scripts. It runs on a configurable interval, evaluates a sequence of conditions, and triggers a system sleep when all conditions agree.

## Use cases

- Suspend after a long download finishes, but only if no other user is active
- Put the machine to sleep once a build or batch job completes
- Sleep at night unless a specific process is still running
- Implement time-based rules (e.g. after midnight) with custom logic in Python

## Running

```sh
# Run as a continuous daemon
auto-suspend run config.yaml

# Evaluate conditions once and exit (useful for testing or one-shot scripts)
auto-suspend run --once config.yaml

# Dry-run: see what the daemon would do without actually suspending
auto-suspend debug config.yaml

# Run with an HTTP API server on the default port (8080)
auto-suspend run --api config.yaml

# Run with the API server on a custom port
auto-suspend run --api --api-port 9090 config.yaml

# Run as a remote decision server for another machine (default port 8080)
auto-suspend run --remote config.yaml

# Run the remote decision server on a custom port
auto-suspend run --remote --remote-port 9090 config.yaml
```

### API mode

When started with `--api`, the daemon also exposes a simple HTTP API. This lets you trigger an immediate suspend from another process or script without waiting for the next daemon cycle.

| Flag | Default | Description |
|---|---|---|
| `--api` | - | Enable the HTTP API server |
| `--api-port` | `8080` | Port the API server listens on |

#### Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/suspend` | Suspend the computer immediately |

```sh
curl -X POST http://localhost:8080/suspend
```

The endpoint uses the same suspend logic as the daemon (respects `AUTOSUSPEND_SUSPEND_COMMAND` if set, otherwise auto-detects per OS) and returns `200 OK` once the suspend command has been dispatched.

### Remote mode

With `--remote`, the daemon does not run on a timer and never suspends the machine it
runs on. Instead it exposes a single HTTP endpoint that evaluates the configured
sequence on demand and reports the verdict as plain text. This lets you dedicate one
machine to deciding whether *another* machine may go to sleep.

`--remote` and `--api` cannot be used together.

| Flag | Default | Description |
|---|---|---|
| `--remote` | - | Run as a remote decision server instead of a daemon |
| `--remote-port` | `8080` | Port the remote server listens on |

#### Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/check` | Run the sequence and answer `yes` or `no` |

```sh
curl http://192.168.0.13:8080/check
# yes
```

The response body is `yes` when the sequence agrees to suspend and `no` when it does
not. If the sequence itself cannot be evaluated (a bad path, an invalid script output,
and so on) the endpoint responds with `500` and the error message as the body.

Requests are handled one at a time, so a slow sequence delays other callers rather than
running several copies of your scripts at once.

### Environment variables

| Variable | Description |
|---|---|
| `AUTOSUSPEND_PYTHON_INTERPRETER_PATH` | Path to the Python interpreter. Defaults to `python3` (or `python` on Windows). |
| `AUTOSUSPEND_SUSPEND_COMMAND` | Override the suspend command (e.g. `systemctl suspend`). Detected automatically per OS if unset. |

## Config file

The config is a YAML file with three top-level fields:

```yaml
run_every: 60          # how often to check, in seconds
master_script: ""      # optional — see below
paths:
  default:
    on_error: ignore
    sequence:
      - script: /path/to/check.py
```

### Scripts

Each script is a Python file that must print exactly `yes` or `no` to stdout:

- `yes` - this condition agrees to suspend
- `no` - this condition does not agree to suspend

The computer is suspended only when the full sequence resolves to `yes`.

### Remote steps

A sequence step can also delegate the decision to another machine running auto-suspend
in [remote mode](#remote-mode). Use `remote` instead of `script` and give it the address
of that machine:

```yaml
run_every: 60

paths:
  default:
    on_error: terminate
    sequence:
      - script: /home/user/scripts/is_anyone_logged_in.py
      - remote: 192.168.0.13:8080
      - script: /home/user/scripts/is_vm_running.py
```

Remote steps take part in the sequence exactly like script steps: they are evaluated in
order, combine with `AND`/`OR` the same way, and support the same `on_error` values
(including a step-level override).

The address may be a bare `host:port`, in which case `http://` and the `/check` path are
filled in automatically. A full URL such as `http://192.168.0.13:8080/check` works too.

For a remote step, an error means the remote could not be reached, responded with an
error status, or returned anything other than `yes` or `no`. Requests time out after 30
seconds.

### Paths

You can define multiple named paths and switch between them dynamically using a `master_script`. The master script is a Python file that prints the name of the path to use on each cycle. If no `master_script` is set, the path named `default` is always used.

```yaml
run_every: 300
master_script: /home/user/scripts/pick_path.py   # prints "weekday" or "weekend"
paths:
  weekday:
    sequence:
      - script: /home/user/scripts/is_after_midnight.py
      - script: /home/user/scripts/no_ssh_sessions.py
  weekend:
    sequence:
      - script: /home/user/scripts/is_after_2am.py
```

### Sequence logic

Steps in a sequence are implicitly combined with `AND`. You can introduce `OR` between individual steps:

```yaml
sequence:
  - script: /scripts/check_a.py
  - OR
  - script: /scripts/check_b.py   # A OR B
  - script: /scripts/check_c.py   # AND C
```

This evaluates as `(A OR B) AND C`.

### Error handling

The `on_error` field controls what happens when a step fails - a script exiting with a non-zero code, or a remote that could not be reached or gave a malformed answer. It can be set at the path level (applies to all steps) or overridden per step:

| Value | Behavior |
|---|---|
| `terminate` | Stop the sequence, do not suspend (default) |
| `ignore` | Treat the step as `no` and continue |
| `treat-as-yes` | Treat the step as `yes` and continue |
| `instant-suspend` | Suspend immediately, skip remaining steps |

```yaml
paths:
  default:
    on_error: ignore          # path-level default
    sequence:
      - script: /scripts/check_a.py
      - script: /scripts/check_b.py
        on_error: treat-as-yes  # step-level override
```

### Full example

```yaml
run_every: 120

paths:
  default:
    on_error: terminate
    sequence:
      - script: /home/user/scripts/is_late_night.py
      - script: /home/user/scripts/download_done.py
        on_error: ignore
      - OR
      - script: /home/user/scripts/battery_low.py
```

This suspends the computer every 120 seconds when:
- it is late night, **and**
- the download is done (errors are ignored and treated as `no`) **or** the battery is low

## Building

Requires Go 1.21+.

```sh
go build -o auto-suspend .
```

The result is a single self-contained binary with no runtime dependencies beyond a Python interpreter on the target machine.

To cross-compile for a specific platform:

```sh
GOOS=linux GOARCH=amd64 go build -o auto-suspend-linux-amd64 .
GOOS=darwin GOARCH=arm64 go build -o auto-suspend-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o auto-suspend-windows-amd64.exe .
```

## Supported platforms

| Platform | Suspend method |
|---|---|
| Linux | `systemctl suspend`, fallback to `pm-suspend` |
| macOS | `pmset sleepnow` |
| Windows | `rundll32.exe powerprof.dll,SetSuspendState` |

You can override the suspend command on any platform via `AUTOSUSPEND_SUSPEND_COMMAND`.
