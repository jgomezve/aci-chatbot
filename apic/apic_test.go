package apic

import (
	"bytes"
	"chatbot/mocks"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestApicClient(t *testing.T) {

	Client = &mocks.MockClient{}
	json := `{
		"totalCount": "1",
		"imdata": [
			{
				"aaaLogin": {
					"attributes": {
						"token": "eyJhbGciOiJSUzI1NiIsImtpZCI6InJqcmRjazBuNW",
						"siteFingerprint": "rjrdck0n5dszr2udbwb6wrdsnwq13qtj",
						"refreshTimeoutSeconds": "600",
						"maximumLifetimeSeconds": "86400",
						"guiIdleTimeoutSeconds": "3600",
						"restTimeoutSeconds": "90",
						"creationTime": "1641149177",
						"firstLoginTime": "1641149177",
						"userName": "jgomezve",
						"remoteUser": "false",
						"unixUserId": "13268",
						"sessionId": "20MjWeChTLebwZxtJu96Cw==",
						"lastName": "Gomez Velasquez",
						"firstName": "Jorge",
						"changePassword": "no",
						"version": "5.2(2f)",
						"buildTime": "Wed Aug 04 17:44:19 UTC 2021",
						"node": "topology/pod-1/node-1"
					}
				}
			}
		]
	}`
	t.Run("Create Raw Client", func(t *testing.T) {

		mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
			r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		}
		clt, _ := NewApicClient("http://mocking.com", "admin", "admin")
		if clt.baseURL != "http://mocking.com" {
			t.Error("error initializing client")
		}
		if clt.tkn == "" {
			t.Error("error getting token")
		}

	})

	t.Run("Create Client with Timeout", func(t *testing.T) {
		mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
			r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		}
		clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
		if clt.baseURL != "http://mocking.com" {
			t.Error("error initializing client")
		}
		if clt.tkn == "" {
			t.Error("error getting token")
		}
		if clt.httpClient.(*mocks.MockClient).Timeout != 8*time.Second {
			t.Error("failed to set timeout")
		}
	})

}
