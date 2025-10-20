package nbody

import (
    "bufio"
    "encoding/json"
    "os"
)

// SnapshotSink: wohin Snapshots geschrieben werden
type SnapshotSink interface {
    OnStart(totalSteps int, snapEvery int) error
    OnSnapshot(tDays float64, bodies []Body) error
    OnEnd(finalTDays float64) error
    Close() error
}

// JSONL writer auf Disk
type JSONLSnapshotWriter struct {
    f     *os.File
    bw    *bufio.Writer
}

type jsonlSnapshot struct {
    TimeDays float64 `json:"time_days"`
    Bodies   []Body  `json:"bodies"`
}

func NewJSONLSnapshotWriter(path string) (*JSONLSnapshotWriter, error) {
    f, err := os.Create(path)
    if err != nil { return nil, err }
    return &JSONLSnapshotWriter{f: f, bw: bufio.NewWriter(f)}, nil
}

func (w *JSONLSnapshotWriter) OnStart(totalSteps int, snapEvery int) error { return nil }

func (w *JSONLSnapshotWriter) OnSnapshot(tDays float64, bodies []Body) error {
    rec := jsonlSnapshot{TimeDays: tDays, Bodies: bodies}
    b, err := json.Marshal(rec)
    if err != nil { return err }
    if _, err := w.bw.Write(b); err != nil { return err }
    if err := w.bw.WriteByte('\n'); err != nil { return err }
    return nil
}

func (w *JSONLSnapshotWriter) OnEnd(finalTDays float64) error { return w.bw.Flush() }
func (w *JSONLSnapshotWriter) Close() error {
    if w.bw != nil { _ = w.bw.Flush() }
    if w.f != nil { return w.f.Close() }
    return nil
}
