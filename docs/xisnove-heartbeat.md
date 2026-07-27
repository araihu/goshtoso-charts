# Xisnove heartbeat brief

Primary user and task: Xisnove administrator or public-status visitor scanning monitor availability before opening monitor detail.

Usage scene and constraints: SSR monitor/status page; visible before JavaScript; status must never rely on color alone; underlying observations may be retained or compacted.

Information priority: monitor state and active incident first, heartbeat second, exact observations/table on detail page when needed.

Xisnove boundary: its raw probe results carry `received_at`, `outcome`, and `latency_ms`; public status currently exposes bounded daily uptime records. Xisnove maps either projection into generic `heartbeat.Point` values. This module does not import Xisnove.

State mapping policy belongs in Xisnove. Intended initial mapping: successful result to `StateUp`; failed result to `StateDown`; application-established degraded state to `StateDegraded`; missing/insufficient evidence to `StateUnknown`.

Visual direction: compact contiguous status marks, semantic CSS tokens, no animation, textual summary, native SVG titles. No panel/card is owned by chart; Xisnove chooses surrounding Goshtoso `panel.Panel` or status-page composition.

No-data policy: render explicit no-data state. Do not present absence of probes as uptime.

Density policy: heartbeat accepts at most 300 points. Xisnove chooses a time bucket suitable for period and resolution before creating points; it must retain the most recent failure signal in that aggregation.
