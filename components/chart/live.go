package chart

// LiveData configures an SSE source for chart types that support full-snapshot
// updates. Event defaults to the standard message event.
type LiveData struct {
	URL   string
	Event string
}

// CartesianSnapshot is the renderer-neutral SSE payload accepted by live Bar
// and Line components. Every series must contain one value per category.
type CartesianSnapshot struct {
	Categories []string                  `json:"categories"`
	Series     []CartesianSnapshotSeries `json:"series"`
}

// CartesianSnapshotSeries is one named series in a categorical live snapshot.
type CartesianSnapshotSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
}
