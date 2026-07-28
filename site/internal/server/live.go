package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	interactive "github.com/araihu/goshtoso-charts/components/interactive"
)

func liveAvailabilityEvents(writer http.ResponseWriter, request *http.Request) {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	if err := writeAvailabilityEvent(writer, availabilitySnapshot(time.Now(), 0)); err != nil {
		return
	}
	flusher.Flush()
	if request.Context().Err() != nil {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	step := 1
	for {
		select {
		case <-request.Context().Done():
			return
		case now := <-ticker.C:
			if err := writeAvailabilityEvent(writer, availabilitySnapshot(now, step)); err != nil {
				return
			}
			flusher.Flush()
			step++
		}
	}
}

func writeAvailabilityEvent(writer http.ResponseWriter, snapshot interactive.CartesianSnapshot) error {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "event: chart\ndata: %s\n\n", payload)
	return err
}

func availabilitySnapshot(now time.Time, step int) interactive.CartesianSnapshot {
	const bucketCount = 36
	const bucketWidth = 2 * time.Second
	categories := make([]string, bucketCount)
	series := []interactive.CartesianSnapshotSeries{
		{Name: "Healthy", Values: make([]float64, bucketCount)},
		{Name: "Degraded", Values: make([]float64, bucketCount)},
		{Name: "Down", Values: make([]float64, bucketCount)},
	}
	end := now.Truncate(bucketWidth)
	for index := range categories {
		bucketTime := end.Add(time.Duration(index-bucketCount+1) * bucketWidth)
		categories[index] = bucketTime.Format("15:04:05")
		state := 0
		switch phase := (step + index) % 24; {
		case phase >= 8 && phase <= 10:
			state = 1
		case phase >= 17 && phase <= 19:
			state = 2
		}
		series[state].Values[index] = 1
	}
	return interactive.CartesianSnapshot{Categories: categories, Series: series}
}
