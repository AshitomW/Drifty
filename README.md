# Drifty - The System Change Tracker

Drifty is a free and open-source tool that helps you keep track of your computer or server. It acts like a security camera for your internal system settings. It takes a detailed picture of your system at one point in time, and then compares it to a picture taken later. This lets you see exactly what has changed, when it changed, and what the difference is.

It is written in the Go programming language, which means it is very fast, run on almost any computer (Linux and Mac), and does not require you to install heavy other software to run it.

## What is this tool and why did I build it?

Computers change all the time. Sometimes you change things on purpose, like installing a new program or updating a password. But frequent changes happen when you do not expect them. Maybe an automatic update changed a setting file, or maybe a coworker changed a configuration and forgot to tell you. In the worst case, a bad actor might have broken into your system and hidden a malicious file.

Drifty is designed to answer a simple question: "Is my computer exactly the same as it was yesterday?"

If the answer is "No", Drifty will give you a list of exactly what is different. It helps you find problems before they crash your server, and it helps you find security breaches before they cause damage.

## How does it actually work?

Drifty works in a very logical way:

1.  **Take a Snapshot**: You run Drifty, and it looks at all your important files, programs, environment variables, and settings. It writes all this detailed information down into a single file called a "snapshot".
2.  **Wait**: Time passes. You continue using your computer.
3.  **Take Another Snapshot**: You run Drifty again to get a new picture of your system.
4.  **Compare**: You tell Drifty to open both snapshot files and find the differences.

It highlights specific things like:

- "This specific file was changed by user 'root'."
- "This program version was updated from 1.0 to 1.1."
- "This background service stopped running."

## Understanding What Drifty Monitors

Drifty looks at many different parts of your system to build a complete picture. Here is a detailed explanation of each part.

### 1. File Changes

This is the most important part of the tool. Drifty looks at specific folders that you tell it to watch, like the `/etc` folder where your system settings live.

It does not just look at the file name. It calculates a "digital fingerprint" (called a hash) for the file contents. If you change even one letter inside a text file, that fingerprint changes entirely. This means Drifty can tell if a file has been tampered with, even if the date and size look the same.

It also checks:

- **Who owns the file**: Did the owner change from "root" to a regular user?
- **Permissions**: Did the file become executable (able to run as a program) when it definitely should not be?
- **Size**: Did the file suddenly get much larger or smaller?

**Note about large files**: To keep things fast and prevent your computer from slowing down, Drifty will only check the size and name of files that are larger than 100 megabytes. It will not read the contents of those massive files to check the fingerprint.

### 2. Installed Programs (Packages)

Your computer has a list of all the software installed on it. Drifty asks your computer for this list.

Drifty is smart enough to talk to many different "package managers" (the tools that install software).

- **Linux**: It checks `dpkg` (for Debian/Ubuntu), `rpm` (for RedHat/CentOS), and `apk` (for Alpine Linux).
- **Mac**: It checks `brew` (Homebrew).
- **Programming Languages**: It checks `npm` (for Node.js global packages), `pip` (for Python packages), and `go` (for Go programs).

It saves the exact version number. So if your web server software updates automatically, Drifty will see that change and report it to you.

### 3. Services (Background Programs)

Services are programs that run in the background, like a web server or a database. Drifty checks two very important things about them:

- **Are they running right now?** It checks if they are active, stopped, or failed.
- **Are they set to start automatically?** It checks if they are "enabled" to start when the computer turns on.

This is very important because sometimes malware tries to set itself to start automatically. Drifty would see that a new service was added to the startup list.

### 4. Network Settings

Drifty looks at how your computer talks to the network.

- **Firewall Rules**: It reads the rules (like iptables) that decide which internet traffic is allowed. If someone opens a hole in your firewall to let traffic in, Drifty will see it.
- **DNS Settings**: It checks which server your computer uses to look up website names.
- **Routes**: It checks the map your computer uses to decide where to send data.

### 5. Docker Containers

If you use Docker to run containers, Drifty can see what is happening there too.

- It lists all the containers currently running.
- It checks which "image" (blueprint) each container is using.
- It sees what ports the container has open to the world.

### 6. Security Certificates

Servers use special files called "certificates" to prove their identity (like the padlock icon in your browser). These certificates expire after a certain time, usually one year.

Drifty looks at these files and can warn you if one of them has changed or is about to expire soon. This prevents your website from going down because someone forgot to renew a certificate.

### 7. Environment Variables

These are invisible settings that tell your programs how to run. Sometimes they contain secrets like passwords or API keys.

Drifty records these settings so you can see if they change. However, it is careful not to record your actual passwords. If it sees a variable named `PASSWORD` or `SECRET` or `KEY`, it will replace the value with `****` so your secrets stay safe in the snapshot file.

## How to Install It

Drifty is a single file, so it is easy to install. You need to build it from the source code.

### Step 1: Install Go

You need the Go programming language installed on your computer. You need version 1.24.5 or higher.

### Step 2: Download the Code

Run this command in your terminal to download the code to your computer:

```bash
git clone https://github.com/AshitomW/Drifty.git
cd Drifty
```

### Step 3: Build the Program

Run this command to create the executable file:

```bash
go build -o drift ./cmd/drift
```

This will create a file named `drift` in the current folder. You can run it by typing `./drift`.

## How to Use It

Here are the main ways you will use Drifty.

### 1. Saving a Snapshot

To take a picture of your system right now, run this command:

```bash
./drift snapshot
```

This will create a file with a name like `snapshot-2024-01-01.json` in your current folder.

If you want to give it a specific name, like "Before Update", use the `--name` and `--file` flags:

```bash
./drift snapshot --name "Before Update" --file before-update.json
```

### 2. Comparing Two Snapshots (History)

Let's say you have a snapshot from yesterday (`yesterday.json`) and one from today (`today.json`). To see what changed, run:

```bash
./drift compare yesterday.json today.json
```

Drifty will print a clear table showing you exactly what is new, what is gone, and what changed between yesterday and today.

### 3. Checking for Changes (The Quick Way)

If you have a "perfect" snapshot (we call this a baseline) and you just want to check if anything has changed right now, you can use the `diff` command.

This command takes a new snapshot in memory and immediately compares it to your baseline file. It does not save the new snapshot to disk.

```bash
./drift diff --baseline my-perfect-baseline.json
```

This is great for automatic daily checks. If nothing changed, it says nothing. If something changed, it will tell you.

### 4. Running in Background (Daemon Mode)

If you do not want to run commands manually, you can tell Drifty to run in the background forever. It will wake up every few minutes, take a snapshot, and check it against your baseline file.

To use this, you must have a baseline file first.

```bash
# Check for changes every 1 hour
./drift daemon --baseline my-baseline.json --interval 1h
```

If it detects no changes, it stays silent. If it detects a change, it will print a warning to the screen.

If you want it to save a report file every time it finds a problem, you can give it an output folder:

```bash
# Save report files to the /var/log/drifty folder if changes are found
./drift daemon --baseline my-baseline.json --interval 10m --output /var/log/drifty
```

This is very useful for servers where you want to be alerted immediately if something changes.

## Configuration File

Drifty uses a settings file to know what to check. By default, it looks for `configs/default.yaml`. You can create your own file and tell Drifty to use it with the `-c` flag.

Here is the complete configuration file with explanations for every setting.

```yaml
# configuration for the collector
collector:
  # FILES: Check these folders for changes
  files:
    enabled: true
    paths:
      - /etc # Watch the system config folder
      - /opt/app/config # Watch your application config
      - /usr/local/bin # Watch for new programs
    exclude_paths:
      # Do not check files that match these patterns
      - ".*\\.log$" # Ignore log files because they always change
      - ".*\\.tmp$" # Ignore temporary files
      - ".*\\.swp$" # Ignore files created by text editors
      - "/etc/mtab" # Ignore system mount lists
    follow_links: false # Should we follow shortcuts to other folders? No.
    max_depth: 10 # How many folders deep should we search?
    hash_algo: sha256 # The math logic used to calculate fingerprints. sha256 is checking.

  # ENVIRONMENT VARIABLES: System settings
  env_vars:
    enabled: true
    include:
      # Only look for variables that start with these words
      - "^APP_.*"
      - "^DB_.*"
      - "^AWS_.*"
    exclude:
      # Ignore variables with these words
      - ".*SESSION_TOKEN.*"
    mask_secrets: true # If true, hide passwords with **** (Very safe)

  # PROCESS ENVIRONMENT: Settings for running programs
  # Note: You usually need to be the root user (administrator) to see these.
  process_env_vars:
    enabled: false
    processes:
      # Only check specific programs
      - node
      - python3
      - nginx
      - java
    max_processes: 20 # Only verify the first 20 processes we find
    mask_secrets: true # Always hide secrets here
    exclude:
      - ".*SECRET.*"

  # PACKAGES: Installed software
  packages:
    enabled: true
    managers:
      # List of package managers to check.
      # Remove the ones you do not use to make it faster.
      - dpkg # For Ubuntu/Debian
      - rpm # For CentOS/RedHat
      - apk # For Alpine Linux
      - npm # For Javascript packages
      - pip # For Python packages
      - go # For Go programs
      - brew # For Mac software

  # NETWORK: Internet settings
  network:
    enabled: true
    interfaces: true # Check IP addresses and network cards
    routes: false # Check the routing map (usually noisy)
    dns: true # Check DNS servers
    firewall_rules: true # Check firewall security rules

  # DOCKER: Container settings
  docker:
    enabled: true
    socket_path: /var/run/docker.sock # Where Docker lives
    containers: true # Check running containers
    images: true # Check downloaded images
    volumes: false # Check storage volumes
    networks: false # Check container networks

  # SYSTEM RESOURCES: CPU and Memory usage
  # This records how busy the computer was when the snapshot was taken.
  system_resources:
    enabled: false # Off by default. Turn on if you want this record.
    cpu: true
    memory: true
    disks: true
    load: true

  # SCHEDULED TASKS: Things that run automatically
  scheduled_tasks:
    enabled: true
    cron_jobs: true # Check standard scheduled tasks
    systemd_timers: true # Check systemd timers
    launchd_jobs: false # Check Mac scheduled tasks

  # CERTIFICATES: Security files
  certificates:
    enabled: true
    paths:
      # Folders to scan for security certificates
      - /etc/ssl/certs
      - /etc/letsencrypt/live
    extensions:
      # File endings that look like certificates
      - .pem
      - .crt
      - .key
    days_threshold: 30 # Warn if expiring in less than 30 days

  # USERS & GROUPS: User accounts
  users_groups:
    enabled: false
    users: true # Check for new user accounts
    groups: true # Check for new groups
    sudo_rules: true # Check who is allowed to be administrator

  # SERVICES: Background programs
  services:
    enabled: true
    init_type: systemd # How your computer manages services (systemd for Linux, launchd for Mac)
    include:
      # Only watch specific services. If this list is empty, it watches ALL services.
      - "nginx"
      - "postgresql"
      - "docker"
      - "sshd"
    exclude:
      - ".*\\.timer$" # Do not check timers here

# SEVERITY RULES
# This tells Drifty what is a "Critical" problem.
# If these change, Drifty will warn you loudly.
severity_rules:
  critical_packages:
    - openssl # Security library
    - nginx # Web server
  critical_services:
    - sshd # Remote access service
    - ufw # Firewall
  critical_files:
    - /etc/passwd # User list
    - /etc/shadow # Password hashes
    - /etc/ssh/sshd_config # Remote access config
  critical_env_vars:
    - DATABASE_URL # Database connection string

# OUTPUT SETTINGS
output:
  format: table # How to print results (table is easiest to read)
  color: true # Use colors

# STORAGE SETTINGS
storage:
  type: file
  path: /var/lib/drift-detector/snapshots # Where to save files
```

## Important Things to Know (Limitations)

Drifty is powerful, but it cannot do everything. Here are some things you should know.

1.  **You often need to be an Administrator**: To check important things like the firewall, Docker, or other users' files, you usually need to run Drifty as the "root" user or use `sudo`. If you run it as a normal user, it might say "Permission Denied" for some checks.
2.  **It does not read large files**: As mentioned before, if a file is bigger than 100 megabytes, Drifty will not check its contents (the hash). It will only check if the file size changed.
3.  **It is not a Time Machine**: Drifty only knows about the exact moment you run the `snapshot` command. If you change a file on Monday and change it back on Tuesday, and you only run Drifty on Wednesday, Drifty will never know that change happened. It only sees the present moment.
4.  **It cannot read encrypted files**: Drifty can tell you that an encrypted file has changed, but it cannot see inside the file to tell you _what_ changed.

## License

This project uses the MIT License. This means it is free to use. You can copy it, change it, and use it for work or personal projects.
