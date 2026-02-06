# nsscache-http

An HTTP server that exposes LDAP user, group, and shadow data for use with [nsscache](https://github.com/google/nsscache). It periodically fetches data from an LDAP directory (e.g., FreeIPA) and serves it over HTTP in formats compatible with nsscache's HTTP source.

## Features

- Queries LDAP for passwd, group, and shadow entries
- In-memory caching with configurable TTL
- Serves data in both flat file format (for nsscache) and JSON
- Health endpoint with cache statistics
- Docker support

## Endpoints

| Endpoint | Format | Description |
|----------|--------|-------------|
| `/passwd` | text/plain | Users in passwd file format |
| `/passwd.json` | application/json | Users as JSON array |
| `/group` | text/plain | Groups in group file format |
| `/group.json` | application/json | Groups as JSON array |
| `/shadow` | text/plain | Shadow entries in shadow file format |
| `/shadow.json` | application/json | Shadow entries as JSON array |
| `/health` | application/json | Cache health and statistics |

## Installation

### From source

```bash
go build -o nsscache-http .
```

### Docker

```bash
docker build -t nsscache-http .
docker run -v /path/to/config.yaml:/etc/nsscache/config.yaml nsscache-http
```

## Configuration

Copy `config.example.yaml` to `config.yaml` and adjust:

```yaml
ldap:
  host: "ldap.example.com"
  port: 636
  use_tls: true
  skip_verify: false
  bind_dn: "cn=readonly,dc=example,dc=com"
  bind_password: "secret"  # Or use env var LDAP_BIND_PASSWORD
  base_dn: "dc=example,dc=com"
  user_filter: "(objectClass=posixAccount)"
  group_filter: "(objectClass=posixGroup)"
  shadow_filter: "(objectClass=shadowAccount)"

cache:
  ttl: 300  # seconds

server:
  listen: ":8080"
```

For FreeIPA, use `cn=accounts,dc=example,dc=com` as the base DN to avoid the compat tree and duplicate entries.

## Usage

```bash
./nsscache-http -config /path/to/config.yaml
```

## Client Setup

See [CLIENT_SETUP.md](CLIENT_SETUP.md) for detailed instructions on configuring client machines to use this service with nsscache and SSH certificate authentication via Vault.

## Architecture

```
┌─────────────────┐     ┌───────────────────┐     ┌──────────────┐
│  nsscache cron  │────>│  nsscache-http    │────>│    LDAP      │
│  (client)       │     │  (this service)   │     │  (FreeIPA)   │
└─────────────────┘     └───────────────────┘     └──────────────┘
```

The service maintains an in-memory cache that refreshes every `cache.ttl` seconds. Clients running nsscache fetch from the HTTP endpoints and populate local `/etc/passwd.cache` and `/etc/group.cache` files.

## License

MIT
