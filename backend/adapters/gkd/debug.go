package gkd

import (
	"context"
	"fmt"
	"time"
)

// DebugProbeResponse is the result of a one-shot upstream GKD fetch.
type DebugProbeResponse struct {
	SensorID    string `json:"sensor_id"`
	DurationMs  int64  `json:"duration_ms"`
	BodyLen     int    `json:"body_len"`
	NumReadings int    `json:"num_readings"`
	Error       string `json:"error,omitempty"`
	ErrorType   string `json:"error_type,omitempty"`
}

// DebugProbe runs a single GKD fetch bypassing the lastAttempt gate so we
// can see exactly what the upstream is returning from this environment.
// TEMPORARY — remove after staging investigation.
//
//encore:api public method=GET path=/gkd/debug
func DebugProbe(ctx context.Context) (*DebugProbeResponse, error) {
	const sensorID = "isar/stegen-16602008"
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	start := time.Now()
	readings, body, err := fetch(ctx, sensorID)
	resp := &DebugProbeResponse{
		SensorID:    sensorID,
		DurationMs:  time.Since(start).Milliseconds(),
		BodyLen:     len(body),
		NumReadings: len(readings),
	}
	if err != nil {
		resp.Error = err.Error()
		resp.ErrorType = fmt.Sprintf("%T", err)
	}
	return resp, nil
}
