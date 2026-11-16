# Health reporting

The connector reads the local health endpoints of the core and publishes a
summary of the gateway's health to the cloud, so the platform can tell a
gateway whose update failed or whose modules crash-loop from a healthy one.
Configuration keys are listed in the [README](../README.md#health-reporter).

## Applies when

You change what the reporter reads, how it derives the state, or the payload it
publishes. The payload is read by the platform: senergy-platform-connector
forwards it to Kafka `gateway_health`, connection-log-worker stores it and
notifies owners, connection-log serves it. A change to field names or state
values reaches all three.

## Sources

| Source | Request | On failure |
|---|---|---|
| module-manager | `GET {HR_MM_BASE_URL}/health/deployments?auxiliary_deployments=true` | `modules` null, `modules_error` set |
| core-manager | `GET {HR_CM_BASE_URL}/health` | `core` null, `core_error` set |
| module-manager | `GET {HR_MM_BASE_URL}/repository-modules?installed=true` | `module_versions` null, `module_versions_error` set |
| module-manager | `GET {HR_MM_BASE_URL}/auto-update/status` | `auto_update` null, `auto_update_error` set |

A core-manager answering `404` or `403` has no health route (installer older
than v0.32.0). That is recorded in `core_error` but does not affect the state,
so such gateways still report their modules.

The module versions are neither part of the state nor of the reasons, and a
change of them does not cause a report. They are read only when a report is
published, for another reason or as a heartbeat, and ride along with it. A
module-manager answering `404` or `403` has no such route for
modules (older release); that is recorded as `module versions not available
(status 404)` in `module_versions_error` and does not affect the state. While
the module-manager runs a repository job, such as its daily refresh, it answers
`503`; the report of that moment carries `unexpected status 503` and `null`,
and the next one has the list again.

The auto-update status is read on every check. The response is passed through
verbatim as `auto_update`, so its content follows the module-manager without a
connector release; only `enabled` and `last_run.failed` are parsed. A
module-manager answering `404` or `403` has no such route (older release); that
is recorded as `auto-update not available (status 404)` in `auto_update_error`.
`503` while the module-manager runs a job is recorded as `unexpected status
503`. A failed read adds no reasons and does not affect the state. A change of
the `auto_update` content alone does not cause a report; it rides along with the
next one.

## Payload

Topic `health` on the cloud broker:

```json
{
  "version": 1,
  "time": "2026-10-27T14:00:00Z",
  "state": 2,
  "reasons": ["update v0.31.1 -> v0.32.0 failed at updating binaries", "module github.com/x/y: unhealthy, 4 restarts"],
  "core": {"...": "core-manager response, verbatim"},
  "core_error": "",
  "modules": {"...": "module-manager response, verbatim"},
  "modules_error": "",
  "module_versions": [
    {"id": "github.com/SENERGY-Platform/mgw-mqtt-dc", "name": "MQTT Device Connector", "source": "<source>", "channel": "stable", "version": "v1.1.0", "next_version": "v1.2.0"}
  ],
  "module_versions_error": "",
  "auto_update": {"...": "module-manager response, verbatim"},
  "auto_update_error": ""
}
```

`module_versions` lists all installed modules sorted by `id`, with the
installed variant's source, channel and version. `next_version` is empty if the
channel offers no other version. If the read fails, `module_versions` is `null`
and `module_versions_error` holds the reason.

`auto_update` is the module-manager's auto-update status (settings and the
outcome of the last run). If the read fails, it is `null` and
`auto_update_error` holds the reason. The field is additive; `version` stays `1`.

The payload stays below 60 KiB: the raw `modules` body is dropped first, then
`core`, then `auto_update`, then `module_versions`, each only if the report is still too large and
each replaced by `"omitted: report too large"` in its error field.
The platform rejects reports above 64 KiB.

## Known limits of `next_version`

- It is only as fresh as the module-manager's repository cache, which is
  refreshed daily from the module-manager's versions with the periodic refresh
  and otherwise only by hand.
- A module installed from a repository the module-manager no longer has never
  gets a `next_version`.
- The module-manager compares versions by string inequality, so a lower version
  in the channel also shows as `next_version`.

## State

| State | When |
|---|---|
| `0` undetermined | an update is running and has not exceeded `HR_UPDATE_RUNNING_TIMEOUT`, or a source failed (or core reports state 0) and nothing unhealthy was found |
| `2` unhealthy | core state 2, a core service not ok, a listed deployment, an update with result `failed`, an update still `running` after the timeout, or an enabled auto-update whose last run has failed modules |
| `1` healthy | otherwise |

A running update reports `0` because stopped services are expected then. A
failed update stays unhealthy until a later update succeeds; the daily update
cron retries it.

An enabled auto-update with failed modules in its last run adds the reason
`auto-update of <id> failed: <error>` per module. Skipped modules never affect
the state, and a disabled auto-update never does, even with an old failed run
stored. The reason stays until a later run no longer lists the module as failed.

Reasons are sorted update first, then core, then auto-update, then modules, and
capped at 20, so the cause survives the cap.

## Sending

A report is published when the state or the reasons change, and otherwise
every `HR_HEARTBEAT_INTERVAL` as a heartbeat. Restart counts and last errors
are part of the reason text but not of the comparison, so a crash-looping
module does not cause a report on every check. A failed publish is retried on
the next check.
