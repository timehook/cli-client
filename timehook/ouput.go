package timehook

import (
	"fmt"
	"time"
)

type output struct {
	checkPoints map[string]bool
}

// Connecting returns human readable string about starts connection
func (std *output) Connecting() string {
	std.checkPoints["connecting"] = true
	return "Connecting to timehook.io"
}

// State returns human readable state string depending on the already string
// returns so it doesn't duplicate meaningful messages
func (std *output) State(state *StateResponse) []string {
	var outputs []string

	switch state.Status {
	case "registered", "awaitingClock", "sendingHttp", "succeeded", "failed", "timeout":
	default:
		return []string{fmt.Sprintf("\nUnknown status %s", state.Status)}
	}

	if !std.checkPoints["scheduled"] {
		std.checkPoints["scheduled"] = true
		outputs = append(outputs, fmt.Sprintf("\n[0s] Webhook scheduled at %s", state.ScheduledAt))
	}

	if state.Status == "awaitingClock" {
		outputs = append(outputs, ".")
	}

	if state.Status != "awaitingClock" && state.Status != "registered" && !std.checkPoints["sendingHttp"] {
		std.checkPoints["sendingHttp"] = true
		sec, _ := sinceSec(state.RegisteredAt, state.SendingHttpAt)
		outputs = append(outputs, fmt.Sprintf("\n[%0.fs] Sending webhook at %s", sec, state.SendingHttpAt))
	}

	if state.Status == "succeeded" && !std.checkPoints["final"] {
		std.checkPoints["final"] = true
		sec, _ := sinceSec(state.RegisteredAt, state.SucceededAt)
		outputs = append(outputs, fmt.Sprintf("\n[%0.fs] Webhook succeeded at %s\n\n", sec, state.SucceededAt))
	}

	if state.Status == "failed" && !std.checkPoints["final"] {
		std.checkPoints["final"] = true
		sec, _ := sinceSec(state.RegisteredAt, state.FailedAt)
		outputs = append(outputs, fmt.Sprintf("\n[%0.fs] Webhook failed at %s\n\n", sec, state.FailedAt))
	}

	if state.Status == "timeout" && !std.checkPoints["final"] {
		std.checkPoints["final"] = true
		sec, _ := sinceSec(state.RegisteredAt, state.FailedAt)
		outputs = append(outputs, fmt.Sprintf("\n[%0.fs] Webhook failed due timeout at %s\n\n", sec, state.FailedAt))
	}

	return outputs
}

// Error returns human readable interpreted error
func (*output) Error(err error) string {
	switch err {
	case errTooManyRequests:
		return "."
	default:
		return fmt.Sprintf("\n[ERROR] %s", err)
	}
}

// sinceSec returns the number of seconds between to and from string dates in
// ISO 8601 format
func sinceSec(from, to string) (float64, error) {
	layout := "2006-01-02T15:04:05+0000"

	fromT, err := time.Parse(layout, from)
	if err != nil {
		return 0, err
	}
	toT, err := time.Parse(layout, to)
	if err != nil {
		return 0, err
	}

	return toT.Sub(fromT).Seconds(), nil
}

func NewOutput() *output {
	return &output{map[string]bool{
		"connecting":  false,
		"scheduled":   false,
		"sendingHttp": false,
		"final":       false,
	}}
}
