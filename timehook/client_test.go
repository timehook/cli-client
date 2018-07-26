package timehook_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/timehook/cli-client/mock"
	"github.com/timehook/cli-client/timehook"
)

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
	proc := client.RegisterAndPoll("https://the-domain.com", `{"foo" : "bar"}`, 5, 1*time.Nanosecond)
	for range proc.C {
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
