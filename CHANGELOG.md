# Changelog

All notable changes to this fork are documented here. This fork is based on
`SierraJC/terraform-provider-coolify` with `coolify_application` resource added
from PR #87 plus subsequent fixes.

## v1.1.5 (2026-05-28)

### Fixed
- **`description` perpetual drift.** Field was `Optional + Computed` without a
  PlanModifier, so when HCL leaves it null but Coolify GET returns a value,
  every plan computes `description = (known after apply)` and shows a
  spurious update. Added `UseStateForUnknownUnlessNullString` modifier
  (same pattern used by `custom_labels`, `git_commit_sha`, etc.) so the
  state value is preserved when HCL is null.

## v1.1.4 (2026-05-28)

### Fixed
- **`semanticEqual` now filters Coolify's `redirect-to-https` middleware line.**
  v1.1.0 only filtered `tls.certresolver=letsencrypt`. Coolify v4 also injects
  `traefik.http.routers.<name>.middlewares=redirect-to-https` on the HTTP
  router of any HTTPS-enabled app. Without filtering, Read refresh sees
  state and Coolify diverge by this line, plan modifier fails to suppress
  the diff, and `tofu plan` shows a perpetual update for apps that haven't
  changed. Adding this rule to `filterCertResolver` makes the comparison
  ignore the auto-injection.

## v1.1.3 (2026-05-28)

### Fixed
- **Update flow: force `custom_labels` to plan value in post-apply state.**
  Even with the v1.1.0 semantic-equality modifier + v1.1.2 `preserveCustomLabels`
  helper, Coolify v4 adds runtime-context lines beyond `tls.certresolver`
  (e.g. `traefik.http.routers.<name>.middlewares=redirect-to-https` on HTTPS
  apps). These rules are too varied to enumerate in the normalize filter, and
  attempting to does not survive future Coolify changes. Update now overrides
  `data.CustomLabels = plan.CustomLabels` after `ReadFromAPI`, so Terraform's
  consistency check passes regardless of Coolify mutation. The Read path
  still uses `preserveCustomLabels` (semantic equality OR explicit user
  change via Coolify UI both flow through normally).

## v1.1.2 (2026-05-28)

### Fixed
- **`docker_compose_raw` rejected on Update for non-dockercompose source types.**
  Coolify v4 GET returns `docker_compose_raw` content for any app whose
  `build_pack=dockercompose`, but Update PATCH rejects it with 422 unless
  `source_type=dockercompose`. New helper `conditionalDockerComposeRaw` omits
  the field from the payload for all other source types. Unblocks adoption of
  `private-github-app` apps that use `build_pack=dockercompose` (the YAML
  comes from `docker_compose_location` in the git repo, not raw content).
- **`custom_labels` "produced inconsistent result after apply" error.**
  After Update, Coolify returns server-mutated labels (base64↔plaintext +
  letsencrypt certresolver injection). The provider previously stored these
  in the post-apply state, but Terraform's consistency check compares state
  to plan and raises a hard error on any difference. New helper
  `preserveCustomLabels` keeps the plan value in state if the API value is
  semantically equal (via the v1.1.0 modifier's normalize logic). Genuine
  user changes via Coolify UI still flow through (semantic-different →
  state updated → next plan shows diff).

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
