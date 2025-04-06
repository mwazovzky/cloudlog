# CloudLog Examples

This directory contains examples showing how to use the CloudLog library.

## Directory Structure

- `basic/` - Basic usage of the CloudLog library
- `formatters/` - Examples of different log formatters
- `context/` - Demonstrates context and job name usage
- `testing/` - Shows how to test code that uses CloudLog

## Environment Setup

Some examples require Grafana Loki credentials. You can set them via environment variables:

```bash
export LOKI_URL="http://your-loki-instance/api/v1/push"
export LOKI_USERNAME="your-username"
export LOKI_AUTH_TOKEN="your-token"
```

Or create a `.env` file in the examples directory:

```
LOKI_URL=http://your-loki-instance/api/v1/push
LOKI_USERNAME=your-username
LOKI_AUTH_TOKEN=your-token
```

If you don't have a Loki instance, don't worry! The examples will automatically fall back to console output.

## Running the Examples

### Basic Example

Shows the fundamental usage of the library:

```bash
cd examples
go run main.go
```

### Formatters Example

Demonstrates the different formatters available:

```bash
cd examples/formatters
go run main.go
```

### Context and Job Names Example

Shows how to use context and job names to structure logs:

```bash
cd examples/context
go run main.go
```

### Testing Example

Demonstrates how to test code that uses CloudLog:

```bash
cd examples/testing
go run main.go
```

## Example Output

The examples that use the console client will print directly to stdout, so you can see the formatted logs.

When running with a real Loki instance, the logs will be sent there instead, and you can view them using the Grafana interface.
