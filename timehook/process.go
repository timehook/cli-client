package timehook

import (
	"fmt"
	"time"
)

type RegisterAnPollProcess struct {
	C         chan string
	succeeded bool
	finished  bool
	status    string
}

// Connect indicates to the process that is connecting
func (p *RegisterAnPollProcess) Connect() {
	if p.status == "not started" {
		p.status = "connecting"
		p.C <- "\nconnecting to timehook.io"
	}
}

// State indicates to the process in which state is
func (p *RegisterAnPollProcess) State(s *StateResponse) {
	switch s.Status {
	case "registered":
		p.Connect()
		p.registered(s)
	case "awaitingClock":
		p.Connect()
		p.registered(s)
		p.awaiting(s)
	case "sendingHttp":
		p.Connect()
		p.registered(s)
		p.sendingHTTP(s)
	case "succeeded":
		p.Connect()
		p.registered(s)
		p.sendingHTTP(s)
		p.success(s)
	case "failed":
		p.Connect()
		p.registered(s)
		p.sendingHTTP(s)
		p.failed(s)
	case "timeout":
		p.Connect()
		p.registered(s)
		p.sendingHTTP(s)
		p.timeout(s)
	default:
		p.unknown(s)
	}

}

func (p *RegisterAnPollProcess) registered(s *StateResponse) {
	if p.status == "connecting" {
		p.C <- fmt.Sprintf("\n[0s] webhook scheduled at %s", s.ScheduledAt)
		p.status = "scheduled"
	}
}

func (p *RegisterAnPollProcess) awaiting(s *StateResponse) {
	p.C <- "."
}

func (p *RegisterAnPollProcess) sendingHTTP(s *StateResponse) {
	if p.status == "scheduled" {
		sec := sinceSec(s.RegisteredAt, s.SendingHttpAt)
		p.C <- fmt.Sprintf("\n[%0.fs] sending webhook at %s", sec, s.SendingHttpAt)
		p.status = "sending"
	}
}

func (p *RegisterAnPollProcess) success(s *StateResponse) {
	sec := sinceSec(s.RegisteredAt, s.SucceededAt)
	p.C <- fmt.Sprintf("\n[%0.fs] webhook succeeded at %s\n\n", sec, s.SucceededAt)
	p.succeeded = true
	p.finish()
}

func (p *RegisterAnPollProcess) failed(s *StateResponse) {
	sec := sinceSec(s.RegisteredAt, s.FailedAt)
	p.C <- fmt.Sprintf("\n[%0.fs] webhook failed at %s\n\n", sec, s.FailedAt)
	p.finish()
}

func (p *RegisterAnPollProcess) timeout(s *StateResponse) {
	sec := sinceSec(s.RegisteredAt, s.FailedAt)
	p.C <- fmt.Sprintf("\n[%0.fs] webhook timeout at %s\n\n", sec, s.FailedAt)
	p.finish()
}

func (p *RegisterAnPollProcess) unknown(s *StateResponse) {
	p.C <- fmt.Sprintf("\n[??s] exit with unexpected status '%s'\n\n", s.Status)
	p.finish()
}

// Error indicates to the process the error found
func (p *RegisterAnPollProcess) Error(err error) {
	switch err {
	case ErrTooManyRequests:
		p.C <- "."
	default:
		p.C <- fmt.Sprintf("[Error] %s", err)
		p.finish()
	}
}

func (p *RegisterAnPollProcess) finish() {
	p.finished = true
	close(p.C)
}

func (p *RegisterAnPollProcess) IsSucceeded() bool { return p.succeeded }
func (p *RegisterAnPollProcess) IsFinished() bool  { return p.finished }

// sinceSec returns the number of seconds between to and from string dates in
// ISO 8601 format
func sinceSec(from, to string) float64 {
	layout := "2006-01-02T15:04:05+0000"

	fromT, err := time.Parse(layout, from)
	if err != nil {
		return 0
	}
	toT, err := time.Parse(layout, to)
	if err != nil {
		return 0
	}

	return toT.Sub(fromT).Seconds()
}

func NewRegisterAnPollProcess() *RegisterAnPollProcess {
	return &RegisterAnPollProcess{
		C:      make(chan string, 10),
		status: "not started",
	}
}
