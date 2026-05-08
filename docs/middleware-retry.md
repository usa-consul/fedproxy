# Retry Middleware

The `Retry` middleware automatically re-issues a request to the upstream handler
when the upstream returns a configurable set of HTTP status codes (e.g. 502, 503,
504). This complements the existing `CircuitBreaker` middleware and improves
resilience against transient upstream failures.

## Configuration

```go
type RetryConfig struct {
    MaxAttempts int           // Total attempts including the first (default: 3)
    Delay       time.Duration // Wait between attempts          (default: 100ms)
    RetryOn     []int         // Status codes that trigger retry (default: 502, 503, 504)
    Logger      *log.Logger   // Optional; logs each retry attempt
}
```

Passing a zero-value `RetryConfig` (except `Delay`, which must be > 0 to avoid a
tight loop in tests) falls back to `DefaultRetryConfig`.

## Usage

```go
import "github.com/yourorg/fedproxy/internal/middleware"

retryCfg := middleware.DefaultRetryConfig
retryCfg.Logger = log.New(os.Stderr, "[retry] ", log.LstdFlags)

handler = middleware.Retry(retryCfg, proxyHandler)
```

## Behaviour

1. The handler is called for the first attempt.
2. If the response status is in `RetryOn` **and** `MaxAttempts` has not been
   reached, the middleware sleeps for `Delay` and calls the handler again.
3. On the final attempt the actual status code is forwarded to the client
   regardless of its value.
4. If the upstream returns a non-retryable status on any attempt the response
   is forwarded immediately without further retries.

## Interaction with CircuitBreaker

Place `Retry` **outside** (before) `CircuitBreaker` in the middleware chain so
that each retry attempt is counted as a separate call by the circuit breaker:

```
Retry → CircuitBreaker → RateLimit → ReverseProxy
```
