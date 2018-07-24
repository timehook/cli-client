package timehook_test

import (
	"reflect"
	"testing"

	"github.com/timehook/cli-client/timehook"
)

func TestStdOutput_State(t *testing.T) {
	tt := []struct {
		name  string
		given *timehook.StateResponse
		state timehook.StateResponse
		want  []string
	}{
		{
			name: "registered",
			state: timehook.StateResponse{
				ID:           "the-id",
				RegisteredAt: "2018-01-29T12:32:25+0000",
				ScheduledAt:  "2018-01-29T12:32:55+0000",
				Status:       "registered",
			},
			want: []string{"\n[0s] Webhook scheduled at 2018-01-29T12:32:55+0000"},
		},
		{
			name: "awaiting",
			state: timehook.StateResponse{
				ID:              "the-id",
				RegisteredAt:    "2018-01-29T12:32:25+0000",
				ScheduledAt:     "2018-01-29T12:32:55+0000",
				AwaitingClockAt: "2018-01-29T12:32:26+0000",
				Status:          "awaitingClock",
			},
			want: []string{
				"\n[0s] Webhook scheduled at 2018-01-29T12:32:55+0000",
				".",
			},
		},
		{
			name: "sending",
			state: timehook.StateResponse{
				ID:              "the-id",
				RegisteredAt:    "2018-01-29T12:32:25+0000",
				ScheduledAt:     "2018-01-29T12:32:55+0000",
				AwaitingClockAt: "2018-01-29T12:32:26+0000",
				SendingHttpAt:   "2018-01-29T12:32:55+0000",
				Status:          "sendingHttp",
			},
			want: []string{
				"\n[0s] Webhook scheduled at 2018-01-29T12:32:55+0000",
				"\n[30s] Sending webhook at 2018-01-29T12:32:55+0000",
			},
		},
		{
			name: "succeeded",
			state: timehook.StateResponse{
				ID:              "the-id",
				RegisteredAt:    "2018-01-29T12:32:25+0000",
				ScheduledAt:     "2018-01-29T12:32:55+0000",
				AwaitingClockAt: "2018-01-29T12:32:26+0000",
				SendingHttpAt:   "2018-01-29T12:32:55+0000",
				SucceededAt:     "2018-01-29T12:32:56+0000",
				Status:          "succeeded",
			},
			want: []string{
				"\n[0s] Webhook scheduled at 2018-01-29T12:32:55+0000",
				"\n[30s] Sending webhook at 2018-01-29T12:32:55+0000",
				"\n[31s] Webhook succeeded at 2018-01-29T12:32:56+0000\n\n",
			},
		},
		{
			name: "failed",
			state: timehook.StateResponse{
				ID:              "the-id",
				RegisteredAt:    "2018-01-29T12:32:25+0000",
				ScheduledAt:     "2018-01-29T12:32:55+0000",
				AwaitingClockAt: "2018-01-29T12:32:26+0000",
				SendingHttpAt:   "2018-01-29T12:32:55+0000",
				FailedAt:        "2018-01-29T12:32:57+0000",
				Status:          "failed",
			},
			want: []string{
				"\n[0s] Webhook scheduled at 2018-01-29T12:32:55+0000",
				"\n[30s] Sending webhook at 2018-01-29T12:32:55+0000",
				"\n[32s] Webhook failed at 2018-01-29T12:32:57+0000\n\n",
			},
		},
		{
			name: "timeout",
			state: timehook.StateResponse{
				ID:              "the-id",
				RegisteredAt:    "2018-01-29T12:32:25+0000",
				ScheduledAt:     "2018-01-29T12:32:55+0000",
				AwaitingClockAt: "2018-01-29T12:32:26+0000",
				SendingHttpAt:   "2018-01-29T12:32:55+0000",
				FailedAt:        "2018-01-29T12:32:57+0000",
				Status:          "timeout",
			},
			want: []string{
				"\n[0s] Webhook scheduled at 2018-01-29T12:32:55+0000",
				"\n[30s] Sending webhook at 2018-01-29T12:32:55+0000",
				"\n[32s] Webhook failed due timeout at 2018-01-29T12:32:57+0000\n\n",
			},
		},
		{
			name: "sendingHTTP from registered",
			given: &timehook.StateResponse{
				ID:           "the-id",
				RegisteredAt: "2018-01-29T12:32:25+0000",
				ScheduledAt:  "2018-01-29T12:32:55+0000",
				Status:       "registered",
			},
			state: timehook.StateResponse{
				ID:              "the-id",
				RegisteredAt:    "2018-01-29T12:32:25+0000",
				ScheduledAt:     "2018-01-29T12:32:55+0000",
				AwaitingClockAt: "2018-01-29T12:32:26+0000",
				SendingHttpAt:   "2018-01-29T12:32:55+0000",
				Status:          "sendingHttp",
			},
			want: []string{
				"\n[30s] Sending webhook at 2018-01-29T12:32:55+0000",
			},
		},
		{
			name: "succeeded from registered",
			given: &timehook.StateResponse{
				ID:           "the-id",
				RegisteredAt: "2018-01-29T12:32:25+0000",
				ScheduledAt:  "2018-01-29T12:32:55+0000",
				Status:       "registered",
			},
			state: timehook.StateResponse{
				ID:              "the-id",
				RegisteredAt:    "2018-01-29T12:32:25+0000",
				ScheduledAt:     "2018-01-29T12:32:55+0000",
				AwaitingClockAt: "2018-01-29T12:32:26+0000",
				SendingHttpAt:   "2018-01-29T12:32:55+0000",
				SucceededAt:     "2018-01-29T12:32:56+0000",
				Status:          "succeeded",
			},
			want: []string{
				"\n[30s] Sending webhook at 2018-01-29T12:32:55+0000",
				"\n[31s] Webhook succeeded at 2018-01-29T12:32:56+0000\n\n",
			},
		},
		{
			name: "unknown",
			state: timehook.StateResponse{
				ID:           "the-id",
				RegisteredAt: "2018-01-29T12:32:25+0000",
				ScheduledAt:  "2018-01-29T12:32:55+0000",
				Status:       "unknown-value",
			},
			want: []string{
				"\nUnknown status unknown-value",
			},
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			stdOutput := timehook.NewOutput()
			if v.given != nil {
				stdOutput.State(v.given)
			}
			got := stdOutput.State(&v.state)
			if !reflect.DeepEqual(got, v.want) {
				t.Errorf("\ngot %s \nwant %s", got, v.want)
			}
		})
	}
}
