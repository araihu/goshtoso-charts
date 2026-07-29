# Xisnove live-availability example brief

Primary user and task: Xisnove administrator or public-status visitor scanning monitor availability before opening monitor detail.

Usage scene and constraints: SSR monitor/status page; visible before JavaScript; status must never rely on color alone; underlying observations may be retained or compacted.

Information priority: monitor state and active incident first, heartbeat second, exact observations/table on detail page when needed.

Xisnove boundary: its raw probe results carry `received_at`, `outcome`, and `latency_ms`; public status currently exposes bounded daily uptime records. Xisnove maps either projection into generic Bar-series values or renderer-neutral `CartesianSnapshot` events. This module does not import Xisnove.

State mapping policy belongs in Xisnove. The example uses named Healthy,
Degraded, and Down series; missing or insufficient evidence must remain explicit
in adjacent application content.

Visual direction: tall, narrow stacked bars, chart palette tokens, textual
summary, and named series. No panel/card is owned by the chart; Xisnove chooses
the surrounding Goshtoso composition.

No-data policy: render explicit no-data state. Do not present absence of probes as uptime.

Density policy: Xisnove chooses a time bucket suitable for period and
resolution before producing Bar values or SSE snapshots; it must retain the
most recent failure signal in that aggregation.
