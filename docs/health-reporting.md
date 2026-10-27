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

A core-manager answering `404` or `403` has no health route (installer older
than v0.32.0). That is recorded in `core_error` but does not affect the state,
so such gateways still report their modules.

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
  "modules_error": ""
}
```

The payload stays below 60 KiB: the raw `modules` body is dropped first, then
`core`, each replaced by `"omitted: report too large"` in its error field.
The platform rejects reports above 64 KiB.

## State

| State | When |
|---|---|
| `0` undetermined | an update is running and has not exceeded `HR_UPDATE_RUNNING_TIMEOUT`, or a source failed (or core reports state 0) and nothing unhealthy was found |
| `2` unhealthy | core state 2, a core service not ok, a listed deployment, an update with result `failed`, or an update still `running` after the timeout |
| `1` healthy | otherwise |

A running update reports `0` because stopped services are expected then. A
failed update stays unhealthy until a later update succeeds; the daily update
cron retries it.

Reasons are sorted update first, then core, then modules, and capped at 20, so
the cause survives the cap.

## Sending

A report is published when the state or the reasons change, and otherwise
every `HR_HEARTBEAT_INTERVAL` as a heartbeat. Restart counts and last errors
are part of the reason text but not of the comparison, so a crash-looping
module does not cause a report on every check. A failed publish is retried on
the next check.
