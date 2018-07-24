// Package timehook executes HTTP request to the Timehook API
package timehook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	registerURL = "https://api.timehook.io/webhooks"
	stateURL    = "https://api.timehook.io/states"
)

// HTTPDoer is the interface that wraps the execution of send HTTP requests to
// API server and return HTTP responses
type HTTPDoer interface {
	Do(req *http.Request) (res *http.Response, e error)
}

// RegisterResponse represents the body response when register a webhook
type RegisterResponse struct {
	ID string `json:"id"`
}

// StateResponse represents the body response when query the state of a webhook
type StateResponse struct {
	ID              string `json:"id"`
	RegisteredAt    string `json:"registeredAt"`
	AwaitingClockAt string `json:"awaitingClockAt"`
	SendingHttpAt   string `json:"sendingHttpAt"`
	FailedAt        string `json:"failedAt"`
	SucceededAt     string `json:"succeededAt"`
	ScheduledAt     string `json:"scheduledAt"`
	Status          string `json:"status"`
}

var (
	errTooManyRequests = errors.New("server responses 429 too many request")
	errUnauthorized    = errors.New("server responses 401 unauthorized request")
)

// client exposes the functions allowed to talk with the Timehook API
type client struct {
	key      string
	httpDoer HTTPDoer
}

// RegisterAndPoll starts a long running process for a single webhook and
// returns three channels: first one for regular messages, second one for
// errors and third one indicates in the overall process was successful.
//
// The process consists in first registers the webhook to be execute on URL
// with the body given with a delay in seconds.
// Second it polls the state until it the webhook finishes or until encounter
// an irrecoverable error.
func (c *client) RegisterAndPoll(URL, body string, delay int, interval time.Duration) (<-chan string, <-chan string, <-chan bool) {
	out := make(chan string)
	errc := make(chan string)
	succ := make(chan bool)

	go func() {
		defer close(out)
		defer close(errc)
		defer close(succ)
		stdout := NewOutput()

		out <- stdout.Connecting()
		rr, err := c.register(URL, body, delay)
		if err != nil {
			errc <- stdout.Error(err)
			succ <- false
			return
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			sr, err := c.state(rr.ID)
			if err != nil {
				errc <- stdout.Error(err)
			} else {
				for _, v := range stdout.State(sr) {
					out <- v
				}
			}

			if isFinal(sr, err) {
				succ <- isSuccess(sr, err)
				break
			}
		}
	}()

	return out, errc, succ
}

// isFinal returns if StateResponse or err is a final state
func isFinal(state *StateResponse, err error) bool {
	if state != nil {
		return state.Status == "failed" || state.Status == "succeeded" || state.Status == "timeout"
	}

	if err == errUnauthorized {
		return true
	}

	return false
}

// isSuccess returns if StateResponse or err is a final state and succeeded
func isSuccess(state *StateResponse, err error) bool {
	return isFinal(state, err) && err == nil && state.Status == "succeeded"
}

// register registers a new webhook and returns a RegisterResponse or error
func (c *client) register(URL, body string, delay int) (*RegisterResponse, error) {
	req, err := http.NewRequest(http.MethodPost, registerURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("can not create new request: %s", err)
	}
	req.Header.Set("X-Webhook", URL)
	req.Header.Set("X-Seconds", strconv.Itoa(delay))

	b, err := c.execute(req, 201)
	if err != nil {
		return nil, err
	}

	var rr RegisterResponse
	if err := json.Unmarshal(b, &rr); err != nil {
		return nil, fmt.Errorf("can not parse response %s: %s", b, err)
	}

	return &rr, nil
}

// state query the webhook identify by ID and returns StateResponse or error
func (c *client) state(ID string) (*StateResponse, error) {
	req, err := http.NewRequest(http.MethodGet, stateURL+"/"+ID, nil)
	if err != nil {
		return nil, fmt.Errorf("can not query state: %s", err)
	}

	b, err := c.execute(req, 200)
	if err != nil {
		return nil, err
	}

	var sr StateResponse
	if err := json.Unmarshal(b, &sr); err != nil {
		return nil, fmt.Errorf("can not parse response %s: %s", b, err)
	}

	return &sr, nil
}

// execute configures common requests parameters, sends the HTTP request and
// returns response body or error
func (c *client) execute(req *http.Request, codeWanted int) ([]byte, error) {
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpDoer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can not execute request: %s", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can not read body: %s", err)
	}

	switch c := resp.StatusCode; {
	case c == 401:
		return nil, errUnauthorized
	case c == 429:
		return nil, errTooManyRequests
	case c != codeWanted:
		return nil, fmt.Errorf("wrong response registering webhook: %s, %s\n", resp.Status, b)
	}

	return b, nil
}

// New returns a new Timehook client given the API key and httpDoer implementation
func New(key string, httpDoer HTTPDoer) *client {
	return &client{key, httpDoer}
}
