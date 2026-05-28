# Changelog

All notable changes to this fork are documented here. This fork is based on
`SierraJC/terraform-provider-coolify` with `coolify_application` resource added
from PR #87 plus subsequent fixes.

## v1.1.1 (2026-05-28)

### Fixed
- `ToAPIUpdate` now uses `expand.StringOrNil` for all string fields instead of
  `expand.String`. Coolify v4's GET endpoint returns empty strings (`""`) for
  source-type-incompatible fields (e.g., `docker_compose_raw=""` on a
  `private-github-app` + dockerfile app). The provider previously preserved
  these empty strings in state and sent them back on Update, triggering
  Coolify validation errors like `{"docker_compose_raw":["This field is not
  allowed."]}` (Laravel rejects the field even with `omitempty` on the JSON
  tag — `omitempty` only skips nil pointers, not pointer-to-empty-string).
  StringOrNil collapses `""` → nil, letting omitempty drop the field from
  the payload. Unblocks Tofu adoption of all source-type variants.

### Behavior change
- HCL `field = ""` (empty string) is now equivalent to `field = null` on
  update — the field is omitted from the API payload, preserving Coolify's
  current value. To explicitly clear a value on Coolify, omit the attribute
  from HCL or set it to `null`. Empty-string-as-clear is no longer supported
  (it never worked correctly anyway due to the above bug).

## v1.1.0 (2026-05-28)

### Added
- `CoolifyLabelsSemanticEqual` plan modifier on `custom_labels`. Recognizes
  Coolify v4's server-side label normalization (base64↔plaintext re-encoding +
  automatic `tls.certresolver=letsencrypt` injection on Traefik routers) as
  semantic no-op. Prevents drift loops; enables safe Tofu adoption of existing
  Coolify applications with file-based TLS setups (e.g., CF Origin Cert via
  Traefik dynamic config).
- Unit test coverage for normalization algorithm (12 table-driven cases +
  isolated helper tests).

### Notes
- `ToAPIUpdate` continues to omit `custom_labels` (carried over from v1.0.3) —
  defensive, avoids unnecessary server-side mutation when only other fields
  change.
- Backward-compatible: no schema breakage. Downgrade NOT supported (state
  written by v1.1.0 remains compatible, but v1.0.x will resume showing drift).

## v1.0.3 (2026-05-28)

### Fixed
- Omit `custom_labels` from `UpdateApplicationByUuidJSONRequestBody`. Coolify
  v4 mutates labels on update; sending them explicitly previously overwrote
  user intent. (Mitigation attempt — Coolify still mutates server-side even
  with payload omission; full fix arrives in v1.1.0.)

## v1.0.2 (2026-05-28)

### Fixed
- Omit `destination_uuid` from update payload (create-only on Coolify API).
- Remove `Computed: true` flag from `destination_uuid` schema (API does not
  return it on read, breaks consistency).

## v1.0.1 (2026-05-28)

### Fixed
- Omit `project_uuid`, `server_uuid`, `environment_name` from update payload.
  These fields are create-only on Coolify v4 and return HTTP 422 when sent in
  an update body.

## v1.0.0 (2026-05-28)

### Added
- `coolify_application` resource (CRUD + import + 6 source types) from
  PR #87 by FaureAlexis against upstream SierraJC/terraform-provider-coolify.
- Fork distribution under registry.wapps.co with GPG-signed releases.
