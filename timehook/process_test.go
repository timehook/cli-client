package timehook_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/timehook/cli-client/timehook"
)

func TestRegisterAnPollProcess(t *testing.T) {
	tt := []struct {
		name          string
		given         func(p *timehook.RegisterAnPollProcess)
		when          func(p *timehook.RegisterAnPollProcess)
		wantMsgs      []string
		wantSucceeded bool
		wantFinished  bool
	}{
		{
			name:          "connecting",
			given:         func(p *timehook.RegisterAnPollProcess) {},
			when:          func(p *timehook.RegisterAnPollProcess) { p.Connect() },
			wantSucceeded: false,
			wantFinished:  false,
			wantMsgs:      []string{"\nconnecting to timehook.io"},
		},
		{
			name:          "registered with no connecting",
			given:         func(p *timehook.RegisterAnPollProcess) {},
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateRegistered()) },
			wantSucceeded: false,
			wantFinished:  false,
			wantMsgs: []string{
				"\nconnecting to timehook.io",
				"\n[0s] webhook scheduled at 2018-01-29T12:32:55+0000",
			},
		},
		{
			name:          "registered",
			given:         func(p *timehook.RegisterAnPollProcess) { p.Connect() },
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateRegistered()) },
			wantSucceeded: false,
			wantFinished:  false,
			wantMsgs:      []string{"\n[0s] webhook scheduled at 2018-01-29T12:32:55+0000"},
		},
		{
			name:          "awaiting",
			given:         func(p *timehook.RegisterAnPollProcess) { p.State(stateRegistered()) },
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateAwaiting()) },
			wantSucceeded: false,
			wantFinished:  false,
			wantMsgs:      []string{"."},
		},
		{
			name:          "awaiting with no connecting",
			given:         func(p *timehook.RegisterAnPollProcess) {},
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateAwaiting()) },
			wantSucceeded: false,
			wantFinished:  false,
			wantMsgs: []string{
				"\nconnecting to timehook.io",
				"\n[0s] webhook scheduled at 2018-01-29T12:32:55+0000",
				".",
			},
		},
		{
			name:          "sending http from registered",
			given:         func(p *timehook.RegisterAnPollProcess) { p.Connect() },
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateSending()) },
			wantSucceeded: false,
			wantFinished:  false,
			wantMsgs: []string{
				"\n[0s] webhook scheduled at 2018-01-29T12:32:55+0000",
				"\n[30s] sending webhook at 2018-01-29T12:32:55+0000",
			},
		},
		{
			name:          "sending http from start",
			given:         func(p *timehook.RegisterAnPollProcess) {},
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateSending()) },
			wantSucceeded: false,
			wantFinished:  false,
			wantMsgs: []string{
				"\nconnecting to timehook.io",
				"\n[0s] webhook scheduled at 2018-01-29T12:32:55+0000",
				"\n[30s] sending webhook at 2018-01-29T12:32:55+0000",
			},
		},
		{
			name:          "finish succeeded from sending",
			given:         func(p *timehook.RegisterAnPollProcess) { p.State(stateSending()) },
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateSucceeded()) },
			wantSucceeded: true,
			wantFinished:  true,
			wantMsgs: []string{
				"\n[31s] webhook succeeded at 2018-01-29T12:32:56+0000\n\n",
			},
		},
		{
			name:          "finish failed from sending",
			given:         func(p *timehook.RegisterAnPollProcess) { p.State(stateSending()) },
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateFailed()) },
			wantSucceeded: false,
			wantFinished:  true,
			wantMsgs: []string{
				"\n[32s] webhook failed at 2018-01-29T12:32:57+0000\n\n",
			},
		},
		{
			name:          "finish timeout from sending",
			given:         func(p *timehook.RegisterAnPollProcess) { p.State(stateSending()) },
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateTimeout()) },
			wantSucceeded: false,
			wantFinished:  true,
			wantMsgs: []string{
				"\n[34s] webhook timeout at 2018-01-29T12:32:59+0000\n\n",
			},
		},
		{
			name:          "finish with unknown state",
			given:         func(p *timehook.RegisterAnPollProcess) {},
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateUnknown()) },
			wantSucceeded: false,
			wantFinished:  true,
			wantMsgs: []string{
				"\n[??s] exit with unexpected status 'unknown-status'\n\n",
			},
		},
		{
			name:          "error 429 too many requests",
			given:         func(p *timehook.RegisterAnPollProcess) { p.State(stateAwaiting()) },
			when:          func(p *timehook.RegisterAnPollProcess) { p.Error(timehook.ErrTooManyRequests) },
			wantSucceeded: false,
			wantFinished:  false,
			wantMsgs: []string{
				".",
			},
		},
		{
			name:          "error Unauthorized too many requests",
			given:         func(p *timehook.RegisterAnPollProcess) { p.State(stateAwaiting()) },
			when:          func(p *timehook.RegisterAnPollProcess) { p.Error(timehook.ErrUnauthorized) },
			wantSucceeded: false,
			wantFinished:  true,
			wantMsgs: []string{
				"[Error] server responses 401 unauthorized request",
			},
		},
		{
			name:          "finish succeeded wrong date from sending ",
			given:         func(p *timehook.RegisterAnPollProcess) { p.State(stateSending()) },
			when:          func(p *timehook.RegisterAnPollProcess) { p.State(stateSucceededWrongDate()) },
			wantSucceeded: true,
			wantFinished:  true,
			wantMsgs: []string{
				"\n[0s] webhook succeeded at 2018-01-29 12:32:56\n\n",
			},
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			p := timehook.NewRegisterAnPollProcess()
			v.given(p)
		purge: // purge messages in channel
			for {
				select {
				case <-p.C:
				case <-time.After(1 * time.Millisecond):
					break purge
				}
			}

			v.when(p)

			// then
			var msgs []string
		loop:
			for {
				select {
				case m, ok := <-p.C:
					if !ok {
						break loop
					}
					msgs = append(msgs, m)
				case <-time.After(1 * time.Millisecond):
					break loop
				}
			}
			if !reflect.DeepEqual(v.wantMsgs, msgs) {
				t.Errorf("wrong msgs slice: \nwant %#v \ngot  %#v", v.wantMsgs, msgs)
			}
			if v.wantSucceeded != p.IsSucceeded() {
				t.Errorf("wrong succeded value, want %v got %v", v.wantSucceeded, p.IsSucceeded())
			}
			if v.wantFinished != p.IsFinished() {
				t.Errorf("wrong finished value, want %v got %v", v.wantFinished, p.IsFinished())
			}
		})
	}
}

func stateRegistered() *timehook.StateResponse {
	return &timehook.StateResponse{
		ID:           "the-id",
		RegisteredAt: "2018-01-29T12:32:25+0000",
		ScheduledAt:  "2018-01-29T12:32:55+0000",
		Status:       "registered",
	}
}

func stateAwaiting() *timehook.StateResponse {
	return &timehook.StateResponse{
		ID:              "the-id",
		RegisteredAt:    "2018-01-29T12:32:25+0000",
		ScheduledAt:     "2018-01-29T12:32:55+0000",
		AwaitingClockAt: "2018-01-29T12:32:26+0000",
		Status:          "awaitingClock",
	}
}

func stateSending() *timehook.StateResponse {
	return &timehook.StateResponse{
		ID:              "the-id",
		RegisteredAt:    "2018-01-29T12:32:25+0000",
		ScheduledAt:     "2018-01-29T12:32:55+0000",
		AwaitingClockAt: "2018-01-29T12:32:26+0000",
		SendingHttpAt:   "2018-01-29T12:32:55+0000",
		Status:          "sendingHttp",
	}
}

func stateSucceeded() *timehook.StateResponse {
	return &timehook.StateResponse{
		ID:              "the-id",
		RegisteredAt:    "2018-01-29T12:32:25+0000",
		ScheduledAt:     "2018-01-29T12:32:55+0000",
		AwaitingClockAt: "2018-01-29T12:32:26+0000",
		SendingHttpAt:   "2018-01-29T12:32:55+0000",
		SucceededAt:     "2018-01-29T12:32:56+0000",
		Status:          "succeeded",
	}
}

func stateFailed() *timehook.StateResponse {
	return &timehook.StateResponse{
		ID:              "the-id",
		RegisteredAt:    "2018-01-29T12:32:25+0000",
		ScheduledAt:     "2018-01-29T12:32:55+0000",
		AwaitingClockAt: "2018-01-29T12:32:26+0000",
		SendingHttpAt:   "2018-01-29T12:32:55+0000",
		FailedAt:        "2018-01-29T12:32:57+0000",
		Status:          "failed",
	}
}

func stateTimeout() *timehook.StateResponse {
	return &timehook.StateResponse{
		ID:              "the-id",
		RegisteredAt:    "2018-01-29T12:32:25+0000",
		ScheduledAt:     "2018-01-29T12:32:55+0000",
		AwaitingClockAt: "2018-01-29T12:32:26+0000",
		SendingHttpAt:   "2018-01-29T12:32:55+0000",
		FailedAt:        "2018-01-29T12:32:59+0000",
		Status:          "timeout",
	}
}

func stateUnknown() *timehook.StateResponse {
	return &timehook.StateResponse{
		ID:     "the-id",
		Status: "unknown-status",
	}
}

func stateSucceededWrongDate() *timehook.StateResponse {
	return &timehook.StateResponse{
		ID:              "the-id",
		RegisteredAt:    "2018-01-29T12:32:25+0000",
		ScheduledAt:     "2018-01-29T12:32:55+0000",
		AwaitingClockAt: "2018-01-29T12:32:26+0000",
		SendingHttpAt:   "2018-01-29T12:32:55+0000",
		SucceededAt:     "2018-01-29 12:32:56",
		Status:          "succeeded",
	}
}
