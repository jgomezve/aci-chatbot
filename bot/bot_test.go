package bot

import (
	"aci-chatbot/webex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// Helper function
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

// Testing the creaion of the Bot
func TestCreateBot(t *testing.T) {
	// Basic errorless Bot Creation
	t.Run("Basic Creation", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		b, err := NewBot(&wmc, nil, "http://test_bot.com")
		equals(t, b.url, "http://test_bot.com")
		equals(t, b.commands["/cpu"].help, "Get APIC CPU Information")
		equals(t, b.commands["/cpu"].regex, "\\/cpu")
		equals(t, err, nil)
	})
	// WebexClient unable to get Bot details
	t.Run("Unavailable Webex Client", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		webex.GetBotDetailsF = func() (webex.WebexPeople, error) {
			return webex.WebexPeople{}, errors.New("Generic Webex Error")
		}
		b, err := NewBot(&wmc, nil, "http://test_bot.com")
		equals(t, b.url, "")
		equals(t, err, errors.New("Generic Webex Error"))

	})
	// Webhooks with same name already exists
	t.Run("Existing Webhook", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		webex.GetWebHooksF = func() ([]webex.WebexWebhook, error) {
			log.Println("Mock: Getting Webhook")
			return []webex.WebexWebhook{{
				Name: "Test-Bot",
			}}, nil
		}
		b, err := NewBot(&wmc, nil, "http://test_bot.com")
		equals(t, b.url, "http://test_bot.com")
		equals(t, err, nil)
	})
	// Error deleting the existing webhook
	t.Run("Existing Webhook - Error Deleting", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		webex.GetWebHooksF = func() ([]webex.WebexWebhook, error) {
			log.Println("Mock: Getting Webhook")
			return []webex.WebexWebhook{{
				Name: "Test-Bot",
			}}, nil
		}
		webex.DeleteWebhookF = func(name, tUrl, id string) error {
			return errors.New("Generic Webex Error")
		}
		b, err := NewBot(&wmc, nil, "http://test_bot.com")
		equals(t, b.url, "")
		equals(t, err, errors.New("Generic Webex Error"))
	})
	// Error creating the new Webhook
	t.Run("Error Creating Webhook", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		webex.CreateWebhookF = func(name, url, resource, event string) error {
			return errors.New("Generic Webex Error")
		}

		b, err := NewBot(&wmc, nil, "http://test_bot.com")
		equals(t, b.url, "")
		equals(t, err, errors.New("Generic Webex Error"))
	})
}

func TestTestHandler(t *testing.T) {

	wbxM := webex.WebexMockClient
	wbxM.SetDefaultFunctions()
	webex.GetBotDetailsF = func() (webex.WebexPeople, error) {
		return webex.WebexPeople{
			NickName: "Test",
		}, nil
	}
	b, _ := NewBot(&wbxM, nil, "http://test_bot.com")
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

	wbxM := webex.WebexMockClient
	wbxM.SetDefaultFunctions()
	b, _ := NewBot(&wbxM, nil, "http://test_bot.com")
	request, _ := http.NewRequest(http.MethodGet, "/about", nil)
	response := httptest.NewRecorder()

	b.router.ServeHTTP(response, request)

	exp, _ := webex.GetBotDetailsF()
	expB, _ := json.Marshal(exp)

	if response.Code != http.StatusOK {
		t.Error("wrong http response")
	}

	if string(expB) != response.Body.String() {
		t.Errorf("wrong http body, got %s\n. It should be %s", string(expB), response.Body.String())
	}
}

func TestAboutMeHandlerError(t *testing.T) {

	wbxM := webex.WebexMockClient
	wbxM.SetDefaultFunctions()
	b, _ := NewBot(&wbxM, nil, "http://test_bot.com")
	// Overwrite mock default behaviour. Provoke an error
	webex.GetBotDetailsF = func() (webex.WebexPeople, error) {
		return webex.WebexPeople{}, errors.New("Webex Timeout")
	}
	request, _ := http.NewRequest(http.MethodGet, "/about", nil)
	response := httptest.NewRecorder()

	b.router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Error("wrong http response")
	}

}
