package bot

import (
	"chatbot/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTestHandler(t *testing.T) {

	wbxM := mocks.WebexMockClient
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
