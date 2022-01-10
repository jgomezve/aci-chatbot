package bot

import (
	"chatbot/mocks"
	"chatbot/webex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTestHandler(t *testing.T) {

	wbxM := mocks.WebexMockClient
	mocks.GetBotDetailsF = func() (webex.WebexPeople, error) {
		return webex.WebexPeople{
			NickName: "Test",
		}, nil
	}
	b := NewBot(&wbxM, nil, "http://test_bot.com")
	request, _ := http.NewRequest(http.MethodGet, "/test", nil)
	response := httptest.NewRecorder()

	b.router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Error("wrong http response")
	}

	if response.Body.String() != "I am alive!" {
		t.Error("wrong http body")
	}
}

func TestAboutMeHandler(t *testing.T) {

	wbxM := mocks.WebexMockClient
	b := NewBot(&wbxM, nil, "http://test_bot.com")
	request, _ := http.NewRequest(http.MethodGet, "/about", nil)
	response := httptest.NewRecorder()

	b.router.ServeHTTP(response, request)

	exp, _ := mocks.GetBotDetailsF()
	expB, _ := json.Marshal(exp)

	if response.Code != http.StatusOK {
		t.Error("wrong http response")
	}

	if string(expB) != response.Body.String() {
		t.Errorf("wrong http body, got %s\n. It should be %s", string(expB), response.Body.String())
	}
}

func TestAboutMeHandlerError(t *testing.T) {

	wbxM := mocks.WebexMockClient
	b := NewBot(&wbxM, nil, "http://test_bot.com")
	// Overwrite mock default behaviour. Provoke an error
	mocks.GetBotDetailsF = func() (webex.WebexPeople, error) {
		return webex.WebexPeople{}, errors.New("Webex Timeout")
	}
	request, _ := http.NewRequest(http.MethodGet, "/about", nil)
	response := httptest.NewRecorder()

	b.router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Error("wrong http response")
	}

}
