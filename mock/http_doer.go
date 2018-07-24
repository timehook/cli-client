package mock

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	register = `{
  "_links": {
    "self": "/webhooks",
    "states": "/states/9e9480a4-271b-4708-993a-064509457a23"
  },
  "id": "9e9480a4-271b-4708-993a-064509457a23"
}`
	stateRegistered = `{
  "id": "9e9480a4-271b-4708-993a-064509457a23",
  "registeredAt": "2018-01-29T12:32:25+0000",
  "scheduledAt": "2018-01-29T12:32:55+0000",
  "status": "registered"
}`
	stateAwaitingClock = `{
  "id": "9e9480a4-271b-4708-993a-064509457a23",
  "registeredAt": "2018-01-29T12:32:25+0000",
  "scheduledAt": "2018-01-29T12:32:55+0000",
  "awaitingClockAt": "2018-01-29T12:32:26+0000",
  "status": "awaitingClock"
}`
	stateSending = `{
  "id": "9e9480a4-271b-4708-993a-064509457a23",
  "registeredAt": "2018-01-29T12:32:25+0000",
  "scheduledAt": "2018-01-29T12:32:55+0000",
  "awaitingClockAt": "2018-01-29T12:32:26+0000",
  "sendingHttpAt": "2018-01-29T12:32:55+0000",
  "status": "sendingHttp"
}`
	stateSucceeded = `{
  "id": "9e9480a4-271b-4708-993a-064509457a23",
  "registeredAt": "2018-01-29T12:32:25+0000",
  "scheduledAt": "2018-01-29T12:32:55+0000",
  "awaitingClockAt": "2018-01-29T12:32:26+0000",
  "sendingHttpAt": "2018-01-29T12:32:55+0000",
  "succeededAt": "2018-01-29T12:32:56+0000",
  "status": "succeeded"
}`
	stateFailed = `{
  "id": "9e9480a4-271b-4708-993a-064509457a23",
  "registeredAt": "2018-01-29T12:32:25+0000",
  "scheduledAt": "2018-01-29T12:32:55+0000",
  "awaitingClockAt": "2018-01-29T12:32:26+0000",
  "sendingHttpAt": "2018-01-29T12:32:55+0000",
  "failedAt": "2018-01-29T12:32:56+0000",
  "status": "failed"
}`
)

type httpDoer struct {
	stack []interface{}
	spies []*http.Request
}

// Do returns the next response in the stack, an *http.Response or an error.
// When the stack exhausted it panics
func (c *httpDoer) Do(req *http.Request) (res *http.Response, e error) {
	if len(c.stack) == 0 {
		panic("no more responses in the stack")
	}

	c.spies = append(c.spies, req)

	var pop interface{}
	pop, c.stack = c.stack[0], c.stack[1:]
	switch pop.(type) {
	case error:
		err := pop.(error)
		return nil, err
	case *http.Response:
		res := pop.(*http.Response)
		return res, nil
	}

	panic(fmt.Sprintf("unknown type %T", pop))
}

// Spies returns all the http.Request given as params
func (c *httpDoer) Spies() []*http.Request {
	return c.spies
}

func RegisteredSuccess() *http.Response {
	resp := makeResponse(register)
	resp.Status = "201 Created"
	resp.StatusCode = 201
	return resp
}

func StateRegistered() *http.Response {
	return makeResponse(stateRegistered)
}

func StateAwaiting() *http.Response {
	return makeResponse(stateAwaitingClock)
}

func StateSending() *http.Response {
	return makeResponse(stateSending)
}

func StateSucceeded() *http.Response {
	return makeResponse(stateSucceeded)
}

func StateFailed() *http.Response {
	return makeResponse(stateFailed)
}

func Unauthorized() *http.Response {
	resp := makeResponse("")
	resp.StatusCode = 401
	resp.Status = "401 Unauthorized"
	return resp
}

func TooManyRequest429() *http.Response {
	resp := makeResponse("")
	resp.StatusCode = 429
	resp.Status = "429 Too Many Requests"
	return resp
}

func makeResponse(body string) *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		Proto:      "1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     map[string][]string{},
	}
}

// HTTPClient returns a mock implementation of HTTPDoer stubbing HTTP
// responses in the stack and recording http.Request received.
// Stack should be *http.Responses or error
func HTTPClient(stack []interface{}) *httpDoer {
	return &httpDoer{stack, make([]*http.Request, 0)}
}
