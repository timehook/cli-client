package timehook_test

import (
	"testing"
	"time"

	"errors"

	"io/ioutil"

	"github.com/timehook/cli-client/mock"
	"github.com/timehook/cli-client/timehook"
)

func TestRegisterAndPoll_Channels(t *testing.T) {
	tt := []struct {
		name          string
		httpResponses []interface{}
		outWanted     int
		errWanted     int
		succWanted    bool
	}{
		{
			name: "happy path",
			httpResponses: []interface{}{
				mock.RegisteredSuccess(),
				mock.StateRegistered(),
				mock.StateAwaiting(),
				mock.StateSending(),
				mock.StateSucceeded(),
			},
			outWanted:  5,
			errWanted:  0,
			succWanted: true,
		},
		{
			name: "succeeded with error querying state",
			httpResponses: []interface{}{
				mock.RegisteredSuccess(),
				mock.StateRegistered(),
				errors.New("some error"),
				mock.StateAwaiting(),
				mock.StateSending(),
				mock.StateSucceeded(),
			},
			outWanted:  5,
			errWanted:  1,
			succWanted: true,
		},
		{
			name: "not succeeded with unauthorized error",
			httpResponses: []interface{}{
				mock.Unauthorized(),
			},
			outWanted:  1,
			errWanted:  1,
			succWanted: false,
		},
		{
			name: "not succeeded with failed webhook",
			httpResponses: []interface{}{
				mock.RegisteredSuccess(),
				mock.StateRegistered(),
				mock.StateAwaiting(),
				mock.StateSending(),
				mock.StateFailed(),
			},
			outWanted:  5,
			errWanted:  0,
			succWanted: false,
		},
		{
			name: "succeeded with 429 too many requests",
			httpResponses: []interface{}{
				mock.RegisteredSuccess(),
				mock.StateRegistered(),
				mock.StateAwaiting(),
				mock.TooManyRequest429(),
				mock.StateSending(),
				mock.StateFailed(),
			},
			outWanted:  5,
			errWanted:  1,
			succWanted: false,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			client := timehook.New("api-key", mock.HTTPClient(v.httpResponses))
			out, errc, succ := client.RegisterAndPoll("https://the-domain.com", `{"foo" : "bar"}`, 5, 1*time.Nanosecond)
			timeC := time.After(1 * time.Second)
			var outN, errN int
		loop:
			for {
				select {
				case <-out:
					outN++
				case <-errc:
					errN++
				case isSucc := <-succ:
					if isSucc != v.succWanted {
						t.Errorf("wrong succ, got %v want %v", isSucc, v.succWanted)
					}
					break loop
				case <-timeC:
					t.Errorf("no message receive in succ channel")
				}
			}

			if outN != v.outWanted {
				t.Errorf("wrong number of out messages, want %d got %d", v.outWanted, outN)
			}
			if errN != v.errWanted {
				t.Errorf("wrong number of err messages, want %d got %d", v.errWanted, errN)
			}
		})
	}
}

func TestRegisterAndPoll_HTTPRequest(t *testing.T) {
	// given
	HTTPClient := mock.HTTPClient([]interface{}{
		mock.RegisteredSuccess(),
		mock.StateRegistered(),
		mock.StateAwaiting(),
		mock.StateSending(),
		mock.StateSucceeded(),
	})
	client := timehook.New("api-key", HTTPClient)

	// when
	out, errc, succ := client.RegisterAndPoll("https://the-domain.com", `{"foo" : "bar"}`, 5, 1*time.Nanosecond)
loop:
	for {
		select {
		case <-out:
		case <-errc:
		case <-succ:
			break loop
		}
	}

	// then register
	first := HTTPClient.Spies()[0]
	if first.URL.String() != "https://api.timehook.io/webhooks" {
		t.Errorf("wrong URL want %s got %s", " https://api.timehook.io/webhooks", first.URL.String())
	}
	if first.Header.Get("Authorization") != "Bearer api-key" {
		t.Errorf("wrong header Authorization want %s got %s", "Bearer api-key", first.Header.Get("Authorization"))
	}
	if first.Header.Get("X-Seconds") != "5" {
		t.Errorf("wrong header X-Seconds want %s got %s", "5", first.Header.Get("X-Seconds"))
	}
	if first.Header.Get("X-Webhook") != "https://the-domain.com" {
		t.Errorf("wrong header X-Webhook want %s got %s", "https://the-domain.com", first.Header.Get("X-Webhook"))
	}
	b, _ := ioutil.ReadAll(first.Body)
	if string(b) != `{"foo" : "bar"}` {
		t.Errorf("wrong body want %s got %s", b, `{"foo" : "bar"}`)
	}

	// then state
	for _, r := range HTTPClient.Spies()[1:] {
		if r.URL.String() != "https://api.timehook.io/states/9e9480a4-271b-4708-993a-064509457a23" {
			t.Errorf("wrong URL want %s got %s", "https://api.timehook.io/states/9e9480a4-271b-4708-993a-064509457a23", r.URL.String())
		}
		if first.Header.Get("Authorization") != "Bearer api-key" {
			t.Errorf("wrong header Authorization want %s got %s", "Bearer api-key", first.Header.Get("Authorization"))
		}
	}
}
