package testhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-bdd/assert"
	"io/ioutil"
	"net/http"
)

type addStepper interface {
	AddStep(definition string, step interface{}) error
}

type testHTTPMethods struct {
	tHTTP TestHTTP
}

type ResponseKey struct{}
type RequestKey struct{}

func Build(addStep addStepper, h httpHandler) TestHTTP {
	thhtp := TestHTTP{
		handler: h,
	}

	testHTTP := testHTTPMethods{tHTTP: thhtp}

	_ = addStep.AddStep(`^I make a (GET|POST|PUT|DELETE|OPTIONS) request to "([^"]*)"$`, testHTTP.makeRequest)
	_ = addStep.AddStep(`^the response code equals (\d+)$`, testHTTP.statusCodeEquals)
	_ = addStep.AddStep(`^the response contains a valid JSON$`, testHTTP.validJSON)
	_ = addStep.AddStep(`^the response is "(.*)"$`, testHTTP.theResponseIs)
	_ = addStep.AddStep(`^the response header "(.*)" equals "(.*)"$`, testHTTP.responseHeaderEquals)
	_ = addStep.AddStep(`^I have a (GET|POST|PUT|DELETE|OPTIONS) request "(.*)"$`, testHTTP.iHaveARequest)
	_ = addStep.AddStep(`^I set request header "(.*)" to "(.*)"$`, testHTTP.iSetRequestSetTo)
	_ = addStep.AddStep(`^I set request body to "([^"]*)"$`, testHTTP.iSetRequestBodyTo)
	_ = addStep.AddStep(`^the request has body "(.*)"$`, testHTTP.theRequestHasBody)
	_ = addStep.AddStep(`^I make the request$`, testHTTP.iMakeRequest)

	return thhtp
}

func (t testHTTPMethods) iSetRequestBodyTo(ctx context.Context, body string) (context.Context, error) {
	r := ctx.Value(RequestKey{}).(*http.Request)
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(body)))
	ctx = context.WithValue(ctx, RequestKey{}, r)
	return ctx, nil
}

func (t testHTTPMethods) iSetRequestSetTo(ctx context.Context, headerName, value string) (context.Context, error) {
	req := ctx.Value(RequestKey{}).(*http.Request)
	req.Header.Add(headerName, value)

	ctx = context.WithValue(ctx, RequestKey{}, req)
	return ctx, nil
}

func (t testHTTPMethods) responseHeaderEquals(ctx context.Context, headerName, expected string) (context.Context, error) {
	resp := ctx.Value(ResponseKey{}).(Response)
	given := resp.Header.Get(headerName)

	return ctx, assert.Equals(expected, given)
}

func (t testHTTPMethods) theRequestHasBody(ctx context.Context, body string) (context.Context, error) {
	req := ctx.Value(RequestKey{}).(*http.Request)
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(body)))
	ctx = context.WithValue(ctx, RequestKey{}, req)
	return ctx, nil
}

func (t testHTTPMethods) iHaveARequest(ctx context.Context, method, url string) (context.Context, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return ctx, err
	}

	ctx = context.WithValue(ctx, RequestKey{}, req)
	return ctx, nil
}

func (t testHTTPMethods) theResponseIs(ctx context.Context, expectedResponse string) (context.Context, error) {
	resp := ctx.Value(ResponseKey{}).(Response)
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	ctx = context.WithValue(ctx, ResponseKey{}, resp)
	if err != nil {
		return ctx, fmt.Errorf("an error while reading the body: %s", err)
	}

	if expectedResponse != string(body) {
		return ctx, errors.New("the body doesn't contain a valid JSON")
	}
	return ctx, nil
}

func (t testHTTPMethods) validJSON(ctx context.Context) (context.Context, error) {
	resp := ctx.Value(ResponseKey{}).(Response)
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	ctx = context.WithValue(ctx, ResponseKey{}, resp)
	if err != nil {
		return ctx, fmt.Errorf("an error while reading the body: %s", err)
	}

	if !json.Valid(body) {
		return ctx, errors.New("the body doesn't contain a valid JSON")
	}
	return ctx, nil
}

func (t testHTTPMethods) iMakeRequest(ctx context.Context) (context.Context, error) {
	req := ctx.Value(RequestKey{}).(*http.Request)
	resp, err := t.tHTTP.MakeRequest(req)
	if err != nil {
		return ctx, err
	}

	ctx = context.WithValue(ctx, ResponseKey{}, resp)
	return ctx, nil
}

func (t testHTTPMethods) makeRequest(ctx context.Context, method, url string) (context.Context, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return ctx, err
	}

	resp, err := t.tHTTP.MakeRequest(req)
	if err != nil {
		return ctx, err
	}

	ctx = context.WithValue(ctx, ResponseKey{}, resp)
	return ctx, nil
}

func (t testHTTPMethods) statusCodeEquals(ctx context.Context, expectedStatus int) (context.Context, error) {
	resp := ctx.Value(ResponseKey{}).(Response)

	if expectedStatus != resp.Code {
		return ctx, fmt.Errorf("expected status code: %d but %d given", expectedStatus, resp.Code)
	}
	return ctx, nil
}
