# GeoBlock Middleware

The `GeoBlock` middleware allows or denies requests based on the geographic
country of the client IP address. It integrates with any external GeoIP lookup
function you provide, keeping the middleware itself dependency-free.

## Configuration

```go
type GeoBlockConfig struct {
    CountryCodes []string              // ISO 3166-1 alpha-2 codes
    Block        bool                  // true = blocklist, false = allowlist
    Lookup       func(ip string) (string, error)
}
```

| Field          | Default      | Description                                      |
|----------------|--------------|--------------------------------------------------|
| `CountryCodes` | `[]`         | List of country codes to block or allow          |
| `Block`        | `true`       | Blocklist mode when true, allowlist when false   |
| `Lookup`       | stub (allow) | Function resolving an IP to a country code       |

## Modes

**Blocklist** (`Block: true`): requests from listed countries receive `403 Forbidden`.

**Allowlist** (`Block: false`): only requests from listed countries are permitted;
all others receive `403 Forbidden`.

## Fail-Open Behaviour

If the `Lookup` function returns an error the request is **allowed** through.
This prevents a GeoIP service outage from taking down the proxy.

## Usage

```go
import maxminddb "github.com/oschwald/maxminddb-golang"

db, _ := maxminddb.Open("GeoLite2-Country.mmdb")

lookup := func(ip string) (string, error) {
    var record struct {
        Country struct {
            ISOCode string `maxminddb:"iso_code"`
        } `maxminddb:"country"`
    }
    err := db.Lookup(net.ParseIP(ip), &record)
    return record.Country.ISOCode, err
}

cfg := middleware.GeoBlockConfig{
    CountryCodes: []string{"CN", "RU", "KP"},
    Block:        true,
    Lookup:       lookup,
}

handler = middleware.GeoBlock(cfg)(handler)
```

## Response Body

Blocked requests return `application/json`:

```json
{"country": "CN", "error": "forbidden"}
```

## Notes

- Country codes are normalised to uppercase; config values are case-insensitive.
- The client IP is extracted from `X-Forwarded-For` (first entry) or `RemoteAddr`.
- Chain `GeoBlock` **before** expensive middleware such as auth or rate-limiting.
