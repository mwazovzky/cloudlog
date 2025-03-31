# Technical Design Document (TDD) for Config

## Overview

The `Config` component is responsible for loading and validating application configuration values. It retrieves values from environment variables and ensures that all required values are present and valid.

---

## Architecture

### 1. **Core Components**

- **`Config` Struct**:
  - Represents the application configuration.
  - Contains fields for all required configuration values.
- **`Load` Function**:
  - Loads configuration values from environment variables.
  - Validates that all required values are present and returns an error if any are missing.

---

## Interfaces and Structs

### 1. **Config**

```go
type Config struct {
    LokiURL       string // The URL of the Loki instance
    LokiUsername  string // The username for Loki authentication
    LokiAuthToken string // The authorization token for Loki
}
```

---

## Methods

### 1. **Load**

- **Purpose**: Loads configuration values from environment variables.
- **Signature**:
  ```go
  func Load() (*Config, error)
  ```
- **Behavior**:
  - Reads the `LOKI_URL`, `LOKI_USERNAME`, and `LOKI_AUTH_TOKEN` environment variables.
  - Validates that all values are non-empty.
  - Returns a `Config` instance if all values are valid, or an error if any value is missing or invalid.

---

## Error Handling

### 1. **Missing Configuration Values**

- If any required configuration value is missing, the `Load` function returns an error indicating which value is missing.

### 2. **Invalid Configuration Values**

- If any configuration value is invalid (e.g., empty string), the `Load` function returns an error indicating the issue.

---

## Extensibility

### 1. **Additional Configuration Values**

- The `Config` struct can be extended to include additional fields as needed.
- The `Load` function can be updated to read and validate additional environment variables.

---

## Testing Strategy

### 1. **Unit Tests**

- Mock the environment variables to simulate various scenarios (e.g., all values present, missing values, invalid values).
- Verify that the `Load` function behaves as expected in each scenario.

### 2. **Integration Tests**

- Use real environment variables to test the `Load` function in a controlled environment.
- Verify that the configuration values are loaded and validated correctly.

---

## Test Cases

### 1. Successful Configuration Load

- **Description**: Verify that the `Load` function successfully loads configuration values when all required environment variables are set.
- **Expected Outcome**: The `Load` function returns a valid `Config` instance without errors.

### 2. Missing Loki URL

- **Description**: Verify that the `Load` function returns an error if the `LOKI_URL` environment variable is missing.
- **Expected Outcome**: The `Load` function returns an error indicating that the `LOKI_URL` value is missing.

### 3. Missing Loki Username

- **Description**: Verify that the `Load` function returns an error if the `LOKI_USERNAME` environment variable is missing.
- **Expected Outcome**: The `Load` function returns an error indicating that the `LOKI_USERNAME` value is missing.

### 4. Missing Loki Authorization Token

- **Description**: Verify that the `Load` function returns an error if the `LOKI_AUTH_TOKEN` environment variable is missing.
- **Expected Outcome**: The `Load` function returns an error indicating that the `LOKI_AUTH_TOKEN` value is missing.

### 5. Empty Configuration Values

- **Description**: Verify that the `Load` function returns an error if any required environment variable is set to an empty string.
- **Expected Outcome**: The `Load` function returns an error indicating which value is empty.

---

## Tools and Frameworks

### 1. **Testing Framework**

- Use `testify` for assertions and mocking.

### 2. **Environment Variable Mocking**

- Use the `os` package to set and unset environment variables during tests.

### 3. **Code Coverage**

- Ensure at least 90% test coverage for the `Config` component.

---

## Example Usage

### Loading Configuration

```go
cfg, err := config.Load()
if err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}
```

### Accessing Configuration Values

```go
fmt.Println("Loki URL:", cfg.LokiURL)
fmt.Println("Loki Username:", cfg.LokiUsername)
fmt.Println("Auth Token:", cfg.LokiAuthToken)
```
