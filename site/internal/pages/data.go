package pages

import (
	"time"

	"github.com/araihu/goshtoso-charts/components/heartbeat"
	"github.com/araihu/goshtoso-charts/components/line"
)

var sampleStart = time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)

func sampleHeartbeat() heartbeat.Config {
	return heartbeat.Config{Label: "Public API availability over seven checks", Caption: "Last seven checks: 5 up, 1 degraded, 1 down.", Points: []heartbeat.Point{
		{At: sampleStart, State: heartbeat.StateUp, Latency: 42 * time.Millisecond},
		{At: sampleStart.Add(time.Minute), State: heartbeat.StateUp, Latency: 47 * time.Millisecond},
		{At: sampleStart.Add(2 * time.Minute), State: heartbeat.StateDegraded, Latency: 900 * time.Millisecond},
		{At: sampleStart.Add(3 * time.Minute), State: heartbeat.StateUp, Latency: 51 * time.Millisecond},
		{At: sampleStart.Add(4 * time.Minute), State: heartbeat.StateDown},
		{At: sampleStart.Add(5 * time.Minute), State: heartbeat.StateUp, Latency: 44 * time.Millisecond},
		{At: sampleStart.Add(6 * time.Minute), State: heartbeat.StateUp, Latency: 46 * time.Millisecond},
	}}
}

func sampleDegradedHeartbeat() heartbeat.Config {
	config := sampleHeartbeat()
	config.Label = "DNS resolution availability"
	config.Caption = "Degraded: elevated response time in two recent checks."
	config.Points[4].State = heartbeat.StateDegraded
	return config
}

func sampleFailedHeartbeat() heartbeat.Config {
	config := sampleHeartbeat()
	config.Label = "Webhook delivery availability"
	config.Caption = "Incident open: two consecutive delivery failures."
	config.Points[5].State = heartbeat.StateDown
	config.Points[6].State = heartbeat.StateDown
	return config
}

func sampleLatency() line.Config {
	return line.Config{
		Label:   "HTTPS monitor latency in milliseconds",
		Caption: "Median latency, last seven checks.",
		Labels:  []string{"08:00", "08:01", "08:02", "08:03", "08:04", "08:05", "08:06"},
		Series:  []line.Series{{Name: "Latency (ms)", Values: []float64{42, 47, 900, 51, 2_000, 44, 46}}},
	}
}

func heartbeatCode() string {
	return `@heartbeat.Heartbeat(heartbeat.Config{
  Label: "Public API availability",
  Points: []heartbeat.Point{
    {At: checkedAt, State: heartbeat.StateUp, Latency: 42*time.Millisecond},
  },
})`
}

func noDataCode() string {
	return `@heartbeat.Heartbeat(heartbeat.Config{Label: "API availability"})`
}

func lineCode() string {
	return `@line.Line(line.Config{
  Label: "HTTPS monitor latency in milliseconds",
  Labels: []string{"08:00", "08:01", "08:02"},
  Series: []line.Series{{Name: "Latency (ms)", Values: []float64{42, 47, 51}}},
})`
}
