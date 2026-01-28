# Drifty — System Drift Collector & Comparator

Drifty is a simple Go tool that takes snapshots of your system's configuration and tells you what has changed. It checks files, installed packages, running services, environment variables, and system information, then compares them to find differences.

## Table of Contents

- [What is Drifty?](#what-is-drifty)
- [What Can It Track?](#what-can-it-track)
- [Where Can We Use It?](#where-can-we-use-it)
- [Where It's Not Suitable](#where-its-not-suitable)
- [Before You Start](#before-you-start)
- [Installation](#installation)
- [How to Configure](#how-to-configure)
- [Using the Tool](#using-the-tool)
- [Commands](#commands)
- [Output Types](#output-types)

## What is Drifty?

Drifty helps you detect when something changes on your system without your knowledge. It does this by:

1. Taking a snapshot of your system's current state (baseline)
2. Taking another snapshot later (current state)
3. Comparing them to show what changed

This is useful for making sure servers stay configured the way you want them to be.

## What Can It Track?

- **Files**: Checks if files have been added, removed, or changed (using file hashes)
- **Packages**: Monitors installed software packages
- **Services**: Checks if services are running or stopped
- **Environment Variables**: Tracks system variables and their values
- **OS Information**: Records system details like hostname and OS version

## Where Can We Use It?

- **Checking if your server is configured correctly**: Create a baseline snapshot of a properly configured server, then check regularly to make sure nothing has changed
- **Compliance and auditing**: Keep a record of what your system looks like at different times
- **After incidents or updates**: Compare your system before and after an incident or software update to see what changed
- **Automation in deployment pipelines**: Make sure deployed servers match the expected configuration
- **Team development setup**: Create a standard environment snapshot and check new developer machines against it

## Where It's Not Suitable

- **Personal computers/laptops**: Not practical for personal machines where you intentionally make frequent changes
- **Real-time monitoring**: This tool takes snapshots at specific times - it doesn't continuously monitor for changes. Use other tools for that
- **Application-specific tracking**: It only looks at system-level changes, not application code or database content
- **Large teams managing many servers**: Would need extra tools to collect snapshots from multiple servers and store them centrally
- **Checking inside encrypted files**: Can only look at file metadata and paths, not encrypted content inside files

## Before You Start

### What You Need

- **Go 1.24.5** or later installed on your machine
- **Linux, macOS** or similar Unix system
- **Admin/root access** might be needed to check protected files and services
- Package managers available: dpkg, yum, pip, npm (depending on your system)
- Init system: systemd, sysvinit, or openrc

### System Checks You Should Know About

- It needs permission to read system files and service information
- On large systems with lots of files, scanning might take a while
- Some features need root access to work properly
- Snapshots can contain sensitive information, so store them securely

## Installation

### Step 1: Clone the Code

```bash
git clone <repo-url> drifty
cd drifty
```

### Step 2: Build the Program

```bash
go build -o bin/drift ./cmd/drift
```

### Step 3: Test It Works

```bash
./bin/drift --help
```

## How to Configure

The tool uses a YAML configuration file to control what it checks. There's a default config in `configs/default.yaml` that you can modify.

### Basic Configuration File

```yaml
collector:
  files:
    enabled: true
    paths:
      - /etc
      - /opt/app/config
    exclude_paths:
      - ".*\\.log$"
      - ".*\\.tmp$"
    max_depth: 10
    hash_algo: sha256

  env_vars:
    enabled: true
    include:
      - "^APP_.*"
      - "^DB_.*"
    exclude:
      - ".*SECRET.*"
    mask_secrets: true

  packages:
    enabled: true
    managers:
      - dpkg
      - pip
      - npm

  services:
    enabled: true
    include:
      - "nginx"
      - "postgresql"
    init_type: systemd

severity_rules:
  critical_packages:
    - kernel
  critical_services:
    - postgresql
  critical_files:
    - /etc/passwd
  critical_env_vars:
    - DATABASE_URL
```

### What Each Section Does

**Files**

- `enabled`: Turn file checking on/off
- `paths`: Which folders to check
- `exclude_paths`: Patterns of files to skip (like .log files)
- `max_depth`: How deep to look into folders
- `hash_algo`: Method to check if files changed (sha256 or md5)

**Environment Variables**

- `enabled`: Turn env var checking on/off
- `include`: Which variables to track (uses patterns)
- `exclude`: Variables to skip
- `mask_secrets`: Hide sensitive values in reports

**Packages**

- `enabled`: Turn package checking on/off
- `managers`: Which package managers to check (dpkg, pip, npm, etc.)

**Services**

- `enabled`: Turn service checking on/off
- `include`: Which services to track
- `init_type`: How your system runs services (systemd is most common)

**Severity Rules**

- Mark important packages, services, files, or variables as "critical"
- When they change, the tool will mark it as a critical change

## Using the Tool

### Basic Commands

All commands use these flags:

- `-c` or `--config`: Path to your config file (optional)
- `-o` or `--output`: How to show results - json, yaml, table, or text (default is table)

## Commands

### 1. `snapshot` - Save Your System's Current State

This command takes a photo of your system right now and saves it to a file.

```bash
# Simple snapshot
drift snapshot

# Save to a file with a name
drift snapshot --name baseline-v1 --file baseline.json

# Use a custom config
drift snapshot -c configs/myconfig.yaml -f snapshot.json
```

### 2. `compare` - Check the Difference Between Two Snapshots

This command takes two saved snapshots and shows you what changed between them.

```bash
# Compare two snapshots
drift compare baseline.json current.json

# See the results as JSON
drift compare baseline.json current.json -o json

# See the results as YAML
drift compare baseline.json current.json -o yaml

# See results in a table (default)
drift compare baseline.json current.json
```

The report tells you:

- What files were added, removed, or changed
- What packages were added or removed
- What services changed
- What environment variables changed
- How important the changes are (Critical, Warning, or Info)

### 3. `diff` - Quick Check: Compare Current System to a Baseline

This command is a shortcut - it takes a new snapshot and immediately compares it to a baseline snapshot in one go.

```bash
# Check if anything changed since your baseline
drift diff --baseline baseline.json

# Use a custom config
drift diff --baseline baseline.json -c configs/myconfig.yaml

# See results as JSON
drift diff --baseline baseline.json -o json
```

**Exit codes:**

- Returns 0 if nothing changed
- Returns 1 if something changed (but not critical)
- Returns 2 if critical changes detected

This is useful in scripts:

```bash
drift diff --baseline baseline.json || {
  echo "System changed!"
  exit 1
}
```

## Output Types

### Table (Default)

```bash
drift compare baseline.json current.json
```

Shows results in a nice formatted table you can read in the terminal.

### JSON

```bash
drift compare baseline.json current.json -o json
```

Shows results as JSON, good for programs to read and process.

### YAML

```bash
drift compare baseline.json current.json -o yaml
```

Shows results as YAML, easier for humans to read than JSON.

### Text

```bash
drift compare baseline.json current.json -o text
```

Shows results as plain text, focusing on important changes.

Typical flags (the CLI's exact flags are defined in `cmd/drift/main.go` — this is a summary):

- `--config` — path to a YAML config file (defaults to `configs/default.yaml`).
- `--baseline` — path to a JSON baseline to compare against (optional). If unspecified, the tool will only produce a snapshot report.
- `--output` — path to write the generated report (stdout by default).
- `--verbose` — enable verbose logging.

Examples

- Generate a snapshot to stdout:

```bash
./bin/drift --config configs/default.yaml
```

- Generate a snapshot and write to a file:

```bash
./bin/drift --config configs/default.yaml --output /tmp/snapshot.json
```

- Compare current snapshot with an existing baseline and write a drift report:

```bash
./bin/drift --config configs/default.yaml --baseline baseline.json --output drift_report.json
```

Produced reports

- The application produces structured JSON objects defined in `internal/models`:
  - `DriftReport` — top-level report containing snapshot data and comparison results.
  - `DriftItem` — individual changed item details (what changed, previous value, current value).
  - `DriftSummary` — aggregate counts, severities, and quick summary.

When to use baseline.json

- Keep a canonical `baseline.json` (a snapshot produced previously by `drift`) for a host or host class. Use it to detect configuration drift over time or after planned changes.

Example workflow

1. Generate a canonical baseline for a golden host:

```bash
./bin/drift --config configs/default.yaml --output baseline.json
```

2. Periodically run `drift` on production hosts and compare against `baseline.json`:

```bash
./bin/drift --config configs/default.yaml --baseline baseline.json --output /var/reports/host1-drift.json
```

3. Inspect `host1-drift.json` for unexpected changes or integrate the report into alerting.

Collector details (internal/collector)

- The `internal/collector` package contains the logic to gather system state:
  - `files.go` — collects file metadata for configured paths; tracks size, permissions, checksum if enabled.
  - `packages.go` — gathers installed package information (platform-dependent).
  - `services.go` — enumerates services and their states where applicable.
  - `envvars.go` — captures environment variables in scope for the process; can be adjusted to sample system-level envs.
  - `osinfo.go` — gathers OS metadata (distro, version, kernel, CPU architecture).

Comparison logic (internal/comparator)

- The comparator reads a baseline snapshot and compares each domain (files, packages, services, env vars) to the current snapshot.
- The comparator reports found differences as `DriftItem`s indicating:
  - `Added` — present in current snapshot but not baseline.
  - `Removed` — present in baseline but missing in current snapshot.
  - `Modified` — present in both but with differing attributes (e.g., checksum, version, state, permissions).

Reporter

- The `reporter` package marshals the `DriftReport` into JSON and writes to the configured output. You can extend or add reporter implementations to produce different output formats (YAML, HTML, or send to remote endpoints).

Where to look in the code

- CLI entrypoint: [cmd/drift/main.go](cmd/drift/main.go)
- Collector implementation: [internal/collector](internal/collector)
- Comparison logic: [internal/comparator/comparator.go](internal/comparator/comparator.go)
- Data models: [internal/models](internal/models)
- Default configuration: [configs/default.yaml](configs/default.yaml)

Development notes

- Use `go mod tidy` to ensure dependencies are in sync.
- Run the binary locally with `go run ./cmd/drift` for fast iteration.
- Unit tests can be added per package — put tests in the same package with `_test.go` files.

Security & permissions

- Some collectors require elevated privileges (root) to read protected files or inspect service managers. Running with least privilege is recommended, but if you need full system snapshots, run on a secure host with appropriate controls.

Extending the tool

- Add new collectors under `internal/collector` and register them in the main collection orchestration.
- Add additional reporter backends by implementing the reporter interface and wiring it in `cmd/drift`.

Troubleshooting

- If some collectors return incomplete data, check process permissions and the target platform support (package/service collection is OS-dependent).
- Use `--verbose` to get logging output for diagnosing collection or comparison failures.

Testing recommendations

- Create fixture snapshots in `testdata/` and write unit tests for the comparator logic to assert expected `DriftItem` outputs.

Contributing

- Fork the repo, make feature branches, and open pull requests with tests and documentation for non-trivial changes.

License

- See LICENSE in the project root for the licensing terms.

Appendix — Example `drift` output (abridged)

```json
{
  "summary": { "added": 2, "removed": 1, "modified": 3 },
  "items": [
    {
      "type": "file",
      "path": "/etc/example.conf",
      "change": "modified",
      "details": { "old_checksum": "...", "new_checksum": "..." }
    },
    {
      "type": "package",
      "name": "nginx",
      "change": "added",
      "details": { "version": "1.21.6" }
    }
  ]
}
```

Contact

- For questions, open an issue in the repository.
