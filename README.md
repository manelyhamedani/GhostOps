# GhostOps

GhostOps is a Go command-line tool for streaming application log analysis. It reads a log file line by line, summarizes valid entries, and detects configured incident patterns over fixed time windows. Reports can be displayed in a human-readable format or exported as JSON.

## Features

- **Streaming log processing:** scans the input file line by line instead of loading the entire file into memory.
- **Log validation:** recognizes `INFO`, `WARN`, `ERROR`, and `FATAL` severity levels and counts malformed lines without stopping the full analysis.
- **Summary statistics:** reports valid and malformed entry totals, counts by severity, and counts by service.
- **Rule-based incident detection:** matches configured severity levels and case-insensitive message keywords.
- **Time-window grouping:** groups matches by service and configured time slice.
- **Two report formats:** human-readable terminal output and an optional indented JSON report.
- **Configurable rules:** loads the time slice and incident rules from a JSON configuration file.

## Requirements

- Go installed and available on your `PATH`.
The implementation uses Go's standard library for its application logic. 

## Build

From the repository root:

```bash
go build -o ghostops .
```

Run the test suite with:

```bash
go test ./...
```

## Usage

```text
ghostops [flags] <logfile>
```

The program expects exactly one log-file path.

### Flags

| Flag | Default | Description |
|---|---|---|
| `-c` | `config.json` | Path to the JSON configuration file |
| `-o` | empty | Path to write the JSON report; when omitted, no JSON file is written |

### Examples

Print the report for a log file using the default configuration:

```bash
./ghostops sample.log
```

Use a different configuration file:

```bash
./ghostops -c config.json sample.log
```

Write a JSON report as well as displaying the terminal report:

```bash
./ghostops -c config.json -o report.json sample.log
```

When `less` starts successfully, the terminal report is displayed through the pager; press `q` to exit. If the pager cannot be started, GhostOps prints a warning to standard error and writes the report directly to standard output.

## Log format

Each input line must follow this format:

```text
RFC3339_TIMESTAMP SEVERITY SERVICE MESSAGE
```

For example:

```text
2026-10-01T00:26:06Z INFO payment-service User successfully logged in
2026-10-02T12:20:47Z FATAL user-service Database connection pool exhausted and crashed system
2026-10-02T12:21:00Z ERROR order-service SQL query failed
```

Supported severity values are:

- `INFO`
- `WARN`
- `ERROR`
- `FATAL`

The timestamp must be valid RFC3339. The parser splits each line into at most four parts, so the message may contain spaces. A line that does not match the expected format, contains an invalid timestamp, or uses an unsupported severity is counted as malformed and skipped for analytics.

## Configuration

GhostOps reads its rules from a JSON file. If `-c` is not supplied, it attempts to load `config.json` from the current working directory.

Example configuration:

```json
{
  "TimeSlice": "5m",
  "Rules": [
    {
      "Name": "DATABASE_CONNECTION_FAILURE",
      "Severity": ["ERROR", "FATAL"],
      "Threshold": 1,
      "Keywords": ["database", "connection"],
      "IncidentSeverity": "HIGH"
    },
    {
      "Name": "WARNING_SPIKE",
      "Severity": ["WARN"],
      "Threshold": 3,
      "Keywords": [],
      "IncidentSeverity": "MEDIUM"
    }
  ]
}
```

### Configuration fields

| Field | Description |
|---|---|
| `TimeSlice` | Positive Go duration used to group events into time windows, such as `1m`, `5m`, or `1h`. If omitted, the configuration loader initializes it to `1m`. |
| `Rules` | List of incident rules to evaluate against valid log entries. |
| `Rules[].Name` | Required unique incident identifier. |
| `Rules[].Severity` | Severity levels that qualify for the rule. |
| `Rules[].Threshold` | Positive count required for an incident to be included in the report. |
| `Rules[].Keywords` | Message keywords; matching is case-insensitive. If the list is empty, the rule does not require a keyword match. |
| `Rules[].IncidentSeverity` | Incident severity: `LOW`, `MEDIUM`, or `HIGH`. |

A rule matches a log entry when its severity is included in the rule's `Severity` list and, if keywords are configured, at least one keyword appears in the message. Matching is case-insensitive. A log entry may match multiple rules.

A configured incident is reported when its match count in a service's time window is greater than or equal to its threshold. The report includes the incident name and severity, affected service, time-window start and end, observed frequency, and configured threshold.

## Reports

### Terminal report

The human-readable report includes:

- Total valid entries.
- Total malformed entries.
- Configured time slice.
- Counts by severity.
- Counts by service.
- Incidents that meet their configured thresholds.

Severity and service summary keys are sorted before display. Incident groups are traversed from maps, so their display order is not guaranteed.

### JSON report

When `-o report.json` is supplied, GhostOps writes an indented JSON object containing a `Summary` section and an `Incidents` array. The summary contains valid and malformed entry counts, the time slice, and counts by severity and service. Each incident contains its name, severity, service, start and end times, frequency, and threshold.

The JSON field names follow the Go struct field names (for example, `Summary`, `ValidEntries`, and `Incidents`).

## Architecture

The implementation separates command-line orchestration, configuration, log processing, analysis, and domain models:

- **`main` package:** parses CLI flags, loads configuration, initializes the analyzer, runs processing, optionally writes JSON, and displays the text report through `less`.
- **`internal/config`:** loads JSON configuration and validates the time slice and incident rule names and thresholds.
- **`internal/logProcessor`:** reads the input incrementally, parses individual lines, marks malformed entries, and passes entries to the analyzer.
- **`internal/model`:** defines log entries, severity levels, incident rules, incident matching, and analytics data structures.
- **`Analyzer`:** tracks totals and counts, buckets entries by service and time slice, and produces text and JSON reports.

The analyzer normalizes timestamps to UTC and truncates them to the configured duration to determine the start of each time bucket. Incident frequencies are tracked separately for each service and time bucket.

## Error handling

Malformed log lines are counted and skipped rather than aborting the entire run. Errors that prevent processing or reporting—such as an unreadable log file, invalid configuration, failure to create the JSON output, or a report-writing error—are returned to the CLI and printed to standard error with a non-zero exit status.

## Tests

Run all package tests:

```bash
go test ./...
```

The current tests cover configuration loading, line reading, parsing valid and malformed entries, log processing, and analyzer recording.

## Limitations and implementation notes

- The default configuration path is relative to the current working directory.
- When `less` is available and starts successfully, the terminal report uses it for paging; otherwise, the report is written directly to standard output.
- Incident groups are reported only when their frequency reaches the rule threshold within a service/time bucket.
- Keyword matching checks whether **any** configured keyword appears in the message; it does not require all keywords to be present.
- The implementation uses the Go `bufio.Scanner` default token-size limit, so exceptionally long individual log lines may cause scanning to fail.
- JSON output is created or truncated at the requested path. Choose an output path that should be overwritten.

## Example output

The exact counts and incidents depend on the input log and configuration. A terminal report has this general shape:

```text
SUMMARY
=======
Valid entries:      7
Malformed entries:  1
Time slice:         5m
BY SEVERITY
-----------
ERROR  1
FATAL  1
INFO   3
WARN   2
BY SERVICE
----------
auth-service          1
notification-service  1
order-service         2
payment-service       2
user-service          1
INCIDENTS
---------
DATABASE_CONNECTION_FAILURE:
- Incident severity: HIGH
- Affected service: user-service
- Time range: 2026-10-02T12:20:00Z -> 2026-10-02T12:25:00Z
- Frequency: 1 (Threshold = 1)
```

This is illustrative output; actual formatting and incident results depend on the input data and configured rules.
