mgw-device-cloud-connector
=======

## Health reporter

Reads the health endpoints of the module manager (`/health/deployments?auxiliary_deployments=true`) and the core manager (`/health`) and publishes a report on the cloud MQTT topic `health` when the state or reasons change, and as a heartbeat otherwise. It is disabled if either base URL is empty.

| Env var | Default | Description |
|---|---|---|
| `HR_MM_BASE_URL` | empty | Module manager base URL, e.g. `http://core-api/module-manager` |
| `HR_CM_BASE_URL` | empty | Core manager base URL, e.g. `http://core-api/core-manager` |
| `HR_CHECK_INTERVAL` | 1m (ns) | Interval of the local checks |
| `HR_HEARTBEAT_INTERVAL` | 5m (ns) | Maximum time between two reports |
| `HR_UPDATE_RUNNING_TIMEOUT` | 1h (ns) | Age after which a core update still running counts as unhealthy; before that the state is undetermined |

States: `0` undetermined, `1` healthy, `2` unhealthy. A core manager without a health route (404 or 403) does not affect the state. Restart counts and last errors are part of the reason text but do not count as a change.
