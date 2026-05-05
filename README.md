# fedproxy

Lightweight reverse proxy with built-in support for federal agency SAML and PIV auth flows.

## Installation

```bash
go install github.com/yourorg/fedproxy@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/fedproxy.git && cd fedproxy && go build ./...
```

## Usage

Create a `config.yaml` file:

```yaml
listen: ":8443"
upstream: "https://internal.agency.gov"
auth:
  saml:
    metadata_url: "https://idp.agency.gov/metadata"
    sp_entity_id: "https://proxy.agency.gov"
  piv:
    ca_cert: "/etc/fedproxy/agency-ca.pem"
    require_piv: true
```

Start the proxy:

```bash
fedproxy --config config.yaml
```

All incoming requests are authenticated via SAML or PIV certificate before being forwarded to the configured upstream. Session tokens are issued as signed JWTs after successful authentication.

## Features

- SAML 2.0 SP-initiated and IdP-initiated flows
- PIV/CAC smart card certificate validation
- TLS termination with mutual TLS support
- Minimal configuration, single binary deployment

## Requirements

- Go 1.21+
- A valid SAML IdP metadata endpoint or PIV-compatible CA certificate

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

## License

MIT © yourorg