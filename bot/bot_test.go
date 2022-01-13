package bot

import (
	"aci-chatbot/webex"
	"bytes"
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
func equals(tb testing.TB, act, exp interface{}) {
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

// Test /test Handler
func TestTestHandler(t *testing.T) {
	// HTTP request to the /test URI
	t.Run("Errorless Request", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		b, _ := NewBot(&wmc, nil, "http://test_bot.com")
		request, _ := http.NewRequest(http.MethodGet, "/test", nil)
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)

		equals(t, response.Code, http.StatusOK)
		equals(t, response.Body.String(), "I am alive!")
	})

}

// Test /about Handler
func TestAboutMeHandler(t *testing.T) {
	// HTTP request to the /test URI without errors
	t.Run("Errorless Request", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		b, _ := NewBot(&wmc, nil, "http://test_bot.com")
		request, _ := http.NewRequest(http.MethodGet, "/about", nil)
		response := httptest.NewRecorder()

		b.router.ServeHTTP(response, request)

		exp, _ := webex.GetBotDetailsF()
		expByte, _ := json.Marshal(exp)

		equals(t, response.Code, http.StatusOK)
		equals(t, response.Body.String(), string(expByte))
		equals(t, response.Header(), http.Header{"Content-Type": []string{"application/json"}})

	})

	t.Run("Error Unavailable Webex Service", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		b, _ := NewBot(&wmc, nil, "http://test_bot.com")
		// Make it fail after creating the Bot
		webex.GetBotDetailsF = func() (webex.WebexPeople, error) {
			return webex.WebexPeople{}, errors.New("Webex Timeout")
		}
		request, _ := http.NewRequest(http.MethodGet, "/about", nil)
		response := httptest.NewRecorder()

		b.router.ServeHTTP(response, request)

		equals(t, response.Code, http.StatusInternalServerError)

	})

}

func TestWebHookHanlder(t *testing.T) {
	// HTTP request to the /test URI without errors
	t.Run("Test /cpu command", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		webex.GetMessagesF = func(roomId string, max int) ([]webex.WebexMessage, error) {
			return []webex.WebexMessage{{Text: "/cpu"}}, nil
		}
		b, _ := NewBot(&wmc, nil, "http://test_bot.com")
		reqB := webex.WebexWebhook{
			Data: &webex.WebexWebhookData{
				RoomId: "AbC13",
			},
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()

		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
	})
}
