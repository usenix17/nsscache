# NSScache Client Setup

This document outlines all changes needed on a client machine to use the nsscache-http service for SSH certificate authentication with Vault-signed certs.

## Prerequisites

Install required packages:

```bash
# Debian/Ubuntu
apt install nsscache libnss-cache
```

## Configuration Files

### 1. /etc/nsscache.conf

```ini
[DEFAULT]
source = http
cache = files
maps = passwd, group
timestamp_dir = /var/lib/nsscache
files_dir = /etc
files_cache_filename_suffix = cache

http_passwd_url = https://nsscache.starnix.net/passwd
http_group_url = https://nsscache.starnix.net/group
```

**Key points:**
- `files_dir = /etc` - libnss-cache is hardcoded to look for `/etc/passwd.cache` and `/etc/group.cache`
- `files_cache_filename_suffix = cache` - no leading dot (otherwise you get `passwd..cache`)

### 2. /etc/nsswitch.conf

Update the passwd and group lines:

```
passwd:         files cache
group:          files cache
```

This checks local `/etc/passwd` first, then falls back to the nsscache files.

### 3. /etc/pam.d/sshd

Add this line **before** `@include common-account`:

```
# Allow nsscache users (high UIDs) without local shadow entry
account sufficient pam_succeed_if.so uid >= 1000000 quiet

# Standard Un*x authorization.
@include common-account
```

**Why:** `pam_unix.so` in common-account checks `/etc/shadow` for password status. Users from nsscache don't have local shadow entries, causing "Access denied by PAM account configuration". This rule permits users with UID >= 1000000 (FreeIPA users) without that check.

### 4. Auto-create home directories (optional but recommended)

In `/etc/pam.d/sshd`, ensure this line exists in the session section:

```
session required pam_mkhomedir.so skel=/etc/skel umask=0022
```

## Initial Setup

```bash
# Create timestamp directory
mkdir -p /var/lib/nsscache

# Run initial sync
nsscache update

# Verify cache files were created
ls -la /etc/passwd.cache /etc/group.cache

# Test user lookup
id sasha
getent passwd sasha
```

## Periodic Sync

Add a cron job to keep the cache updated:

```bash
echo '*/10 * * * * root /usr/bin/nsscache update' > /etc/cron.d/nsscache
```

Or create a systemd timer for more control.

## Troubleshooting

### "duplicate key detected" warnings
The HTTP source is returning duplicate entries. Check the LDAP base_dn - for FreeIPA, use `cn=accounts,dc=example,dc=com` to avoid the compat tree.

### `id username` returns "no such user"
1. Check cache files exist: `ls /etc/passwd.cache`
2. Check nsswitch.conf has `cache` in passwd line
3. Verify libnss-cache is installed: `dpkg -l | grep libnss-cache`

### SSH fails with "Access denied for user by PAM account configuration"
The PAM account check is failing. Check:
1. User exists in cache: `id username`
2. PAM rule added before common-account in `/etc/pam.d/sshd`

### Cache files have wrong names (double dots)
Change `files_cache_filename_suffix = cache` (not `.cache`)

## Vault Configuration

These changes lock down SSH cert signing so users can only get certs for their own username.

### 1. Update OIDC role to use preferred_username

```bash
vault write auth/oidc/role/authentik-admin \
  bound_audiences="<CLIENT_ID>" \
  allowed_redirect_uris="https://vault.starnix.net/ui/vault/auth/oidc/oidc/callback,http://localhost:8250/oidc/callback" \
  user_claim="preferred_username" \
  oidc_scopes="openid,profile,email" \
  policies="default,admin-policy" \
  ttl="1h"
```

**Key:** `user_claim="preferred_username"` ensures the alias name is just `sasha`, not `sasha@starnix.net`.

### 2. Get the OIDC mount accessor

```bash
vault auth list -format=json | jq -r '.["oidc/"].accessor'
# Returns something like: auth_oidc_e4166ac3
```

### 3. Update SSH signing role to use templating

```bash
vault write ssh-client-signer/roles/default-user \
  key_type=ca \
  allow_user_certificates=true \
  allowed_users_template=true \
  allowed_users="{{identity.entity.aliases.auth_oidc_e4166ac3.name}}" \
  default_user="{{identity.entity.aliases.auth_oidc_e4166ac3.name}}" \
  ttl=8h \
  max_ttl=24h \
  allowed_extensions="permit-pty" \
  algorithm_signer=rsa-sha2-256
```

Replace `auth_oidc_e4166ac3` with your actual accessor from step 2.

**How it works:**
- `allowed_users_template=true` enables Go templating
- `{{identity.entity.aliases.<accessor>.name}}` resolves to the authenticated user's OIDC username
- User `sasha` can only sign certs for `sasha`, not for `bair` or anyone else

### 4. Delete existing entities (one-time cleanup)

After changing `user_claim`, delete old entities so new ones are created with correct alias names:

```bash
# Find entity ID
vault token lookup -format=json | jq -r '.data.entity_id'

# Delete it
vault delete identity/entity/id/<entity_id>

# Revoke token and re-login
vault token revoke -self
gcert
```

### 5. (Optional) Create a breakglass role for root access

```bash
vault write ssh-client-signer/roles/breakglass \
  key_type=ca \
  allow_user_certificates=true \
  allowed_users="root" \
  ttl=1h \
  max_ttl=1h \
  allowed_extensions="permit-pty" \
  algorithm_signer=rsa-sha2-256
```

Restrict access to this role via Vault policies and add audit alerting.

## Architecture Overview

```
User SSH with Vault cert
        │
        ▼
┌───────────────┐
│  SSH Server   │ ◄── Validates cert against Vault CA
└───────┬───────┘
        │
        ▼
┌───────────────┐
│     NSS       │ ◄── Looks up user in /etc/passwd, then /etc/passwd.cache
└───────┬───────┘
        │
        ▼
┌───────────────┐
│     PAM       │ ◄── Account check (skipped for UID >= 1000000)
└───────┬───────┘
        │
        ▼
    Session created
```

The nsscache files are populated by periodic sync from the HTTP service:

```
┌─────────────────┐     ┌───────────────────┐     ┌──────────────┐
│  nsscache cron  │────▶│ nsscache.starnix  │────▶│   FreeIPA    │
│  (every 10 min) │     │   .net (HTTP)     │     │    (LDAP)    │
└─────────────────┘     └───────────────────┘     └──────────────┘
```
