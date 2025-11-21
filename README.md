# Renovates

A Go-based wrapper for the [Renovate](https://github.com/renovatebot/renovate) CLI. This tool automates the process of running Renovate against multiple repositories, parsing the output, and sending notifications about dependency updates.

It is designed to run in a "Dry Run" mode to detect updates without automatically creating Pull Requests, making it ideal for reporting and monitoring purposes.

## Features

- **Multi-Platform Support**: Discover repositories on GitHub and GitLab.
- **Auto-Discovery**: Automatically find repositories based on owner, topics, and regex patterns (includes/excludes).
- **Concurrent Execution**: Run Renovate on multiple repositories in parallel to save time.
- **Advanced Log Parsing**: Parses Renovate's JSON logs to extract detailed update information (package name, version changes, file paths, update types).
- **Flexible Notifications**:
  - **Stdout**: Print updates to the console.
  - **Webhook**: Send JSON payloads to a generic webhook URL.
  - **Microsoft Teams**: Send formatted Adaptive Cards (via Power Automate or Incoming Webhook).
  - **Telegram**: Send Markdown-formatted messages via Telegram Bot.

## Prerequisites

- **Go**: 1.25+
- **Renovate CLI**: Must be installed and available in the system PATH.
  ```bash
  npm install -g renovate
  ```

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/snowmerak/renovates.git
   cd renovates
   ```

2. Install Go dependencies:
   ```bash
   go mod tidy
   ```

## Configuration

1. Copy the example configuration file:
   ```bash
   cp config.example.toml config.toml
   ```

2. Edit `config.toml` with your settings:

   ```toml
   command = "renovate"
   platform = "github" # or "gitlab"
   token = "YOUR_PLATFORM_TOKEN"
   endpoint = "https://api.github.com" # or your GitLab instance URL
   concurrency = 5 # Number of concurrent renovations

   # Repository Discovery Settings
   [discovery]
   enabled = true
   owner = "your-org-or-username"
   topics = ["renovate-enabled"] # Optional: filter by topic
   includes = ["^service-.*"]    # Optional: regex to include repos
   excludes = [".*-deprecated$"] # Optional: regex to exclude repos

   # Notification Settings
   [[notifiers]]
   type = "stdout"

   [[notifiers]]
   type = "telegram"
   token = "YOUR_BOT_TOKEN"
   chat_id = "YOUR_CHAT_ID"
   ```

## Usage

### Single Repository
Run Renovate on a specific repository:
```bash
go run . owner/repository-name
```

### Auto-Discovery Mode
Run Renovate on all matching repositories defined in `config.toml`:
```bash
go run .
```
*Note: Ensure `[discovery] enabled = true` is set in your config.*

## Notifications

### Microsoft Teams
Sends an Adaptive Card with a summary of updates.
```toml
[[notifiers]]
type = "teams"
url = "YOUR_TEAMS_WEBHOOK_URL"
```

### Telegram
Sends a Markdown-formatted message.
```toml
[[notifiers]]
type = "telegram"
token = "YOUR_BOT_TOKEN"
chat_id = "YOUR_CHAT_ID"
```

### Generic Webhook
Sends a JSON payload containing the list of updates.
```toml
[[notifiers]]
type = "webhook"
url = "YOUR_WEBHOOK_URL"
```

**Payload Structure:**
```json
{
  "repo": "owner/repository-name",
  "updates": [
    {
      "depName": "github.com/pkg/errors",
      "currentVersion": "v0.9.0",
      "newVersion": "v0.9.1",
      "updateType": "patch",
      "packageFile": "go.mod"
    }
  ]
}
```

## License

AGPL-3.0
