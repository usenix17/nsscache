# Vault OIDC External Group Mapping

## Overview

Vault uses **external groups** to map identity provider (OIDC) groups to Vault policies. This allows you to manage access control in your identity provider (Authentik/FreeIPA) while Vault enforces the policies.

## Two Separate Names

When setting up external group mapping, there are **two different names** that serve different purposes:

### 1. Vault Internal Group Name

This is Vault's internal identifier for the group. It's for your reference in Vault.
```bash
vault write identity/group name="root-access-users" \
  type="external" \
  policies="ssh-root-role"
```

- **Purpose**: Human-readable name within Vault
- **Used for**: Vault CLI/UI, organizing groups internally
- **Not sent to or from OIDC**

### 2. Group Alias Name

This is the **external identifier** that Vault matches against the OIDC `groups` claim.
```bash
vault write identity/group-alias \
  name="breakglass" \
  mount_accessor="${OIDC_ACCESSOR}" \
  canonical_id="${GROUP_ID}"
```

- **Purpose**: Links external identity (OIDC group) to internal Vault group
- **Must match**: The exact value in the OIDC token's `groups` claim
- **Case-sensitive**: "breakglass" ≠ "BreakGlass"

## The Complete Flow

1. **FreeIPA**: Group "breakglass" with members sasha, bair, rslarson
2. **LDAP Sync**: Authentik syncs the group from FreeIPA
3. **Authentik**: Group "breakglass" now exists in Authentik with same members
4. **OIDC Authentication**: User logs in via OIDC
5. **OIDC Token**: Authentik sends ID token with `groups: ["breakglass", "admins", "wheel"]`
6. **Vault Reads Token**: Vault extracts the `groups` claim from the token
7. **Group Alias Matching**: Vault checks if "breakglass" matches any group alias name
8. **Alias Found**: Group alias "breakglass" points to internal group ID `473048b6-6272-417b-40ae-197523b3d269`
9. **Internal Group**: Vault looks up the internal group, finds it has policy `ssh-root-role`
10. **Policy Applied**: User token gets `identity_policies: ["ssh-root-role"]`
11. **Access Granted**: User can now sign SSH certificates with root-role

## Example: Names Can Differ

You can use different names internally vs. externally:
```bash
# Create Vault internal group with descriptive name
vault write identity/group name="root-access-users" \
  type="external" \
  policies="ssh-root-role"

# Get the group ID
GROUP_ID=$(vault read -field=id identity/group/name/root-access-users)

# Create alias that maps to OIDC group name "breakglass"
vault write identity/group-alias \
  name="breakglass" \
  mount_accessor="${OIDC_ACCESSOR}" \
  canonical_id="${GROUP_ID}"
```

**Result:**
- OIDC sends: `groups: ["breakglass"]`
- Vault sees "breakglass" in the groups claim
- Maps to internal group "root-access-users" via the alias
- Applies policies from "root-access-users" group
- User gets `ssh-root-role` policy

## Key Takeaway

The **group alias name** is what matters for matching. It must exactly match the value sent in the OIDC `groups` claim. The **internal group name** is just for organization within Vault and doesn't need to match anything external.
