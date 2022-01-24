package bot

import (
	"aci-chatbot/apic"
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
		equals(t, b.commands["/cpu"].help, "Get APIC CPU Information 💾")
		equals(t, b.commands["/cpu"].regex, "\\/cpu$")
		equals(t, err, nil)
	})
	// WebexClient unable to get Bot details
	t.Run("Unavailable Webex Client", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		wmc.GetBotDetailsF = func() (webex.WebexPeople, error) {
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
		wmc.GetWebHooksF = func() ([]webex.WebexWebhook, error) {
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
		wmc.GetWebHooksF = func() ([]webex.WebexWebhook, error) {
			log.Println("Mock: Getting Webhook")
			return []webex.WebexWebhook{{
				Name: "Test-Bot",
			}}, nil
		}
		wmc.DeleteWebhookF = func(name, tUrl, id string) error {
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
		wmc.CreateWebhookF = func(name, url, resource, event string) error {
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

		exp, _ := wmc.GetBotDetailsF()
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
		wmc.GetBotDetailsF = func() (webex.WebexPeople, error) {
			return webex.WebexPeople{}, errors.New("Webex Timeout")
		}
		request, _ := http.NewRequest(http.MethodGet, "/about", nil)
		response := httptest.NewRecorder()

		b.router.ServeHTTP(response, request)

		equals(t, response.Code, http.StatusInternalServerError)

	})

}

// Test Webhook Handler
func TestWebHookHanlderGeneral(t *testing.T) {

	// Invalid Payload in the POST Request on /webhok
	t.Run("Invalid Webhook Payload", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		amc := apic.ApicMockClient
		amc.SetDefaultFunctions()
		b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
		// Make it fail by sending another payload
		reqB := webex.WebexMessage{
			Markdown: "DummyTest",
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()

		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusInternalServerError)
		equals(t, wmc.LastMsgSent, "")
	})
	// Error Reading Message from Webex
	t.Run("Error Webex Message", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		amc := apic.ApicMockClient
		amc.SetDefaultFunctions()
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{}, errors.New("Generic Webex Error")
		}
		b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
		reqB := webex.WebexWebhook{
			Name: "test-bot",
			Data: &webex.WebexWebhookData{
				RoomId: "AbC13",
			},
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()

		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusInternalServerError)
		equals(t, wmc.LastMsgSent, "")
	})
	// Message which triggered the Webhook came from the bot itself
	t.Run("Message from Bot", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		amc := apic.ApicMockClient
		amc.SetDefaultFunctions()
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/cpu", PersonId: "BotId"}, nil
		}
		wmc.GetBotDetailsF = func() (webex.WebexPeople, error) {
			return webex.WebexPeople{Id: "BotId"}, nil
		}
		b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
		reqB := webex.WebexWebhook{
			Name: "test-bot",
			Data: &webex.WebexWebhookData{
				RoomId: "AbC13",
			},
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusAccepted)

	})
}

func TestWebHookHanlderCpuCommand(t *testing.T) {
	// Test a /CPU command/message without errors
	t.Run("Errorless /cpu command", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		amc := apic.ApicMockClient
		amc.SetDefaultFunctions()
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/cpu"}, nil
		}
		b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
		reqB := webex.WebexWebhook{
			Name: "test-bot",
			Data: &webex.WebexWebhookData{
				RoomId: "AbC13",
			},
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()

		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n\t\nThis is the CPU information of the controllers: \n\n" +
			"<ul><li><code>APIC 1</code> -> \t💻 <strong>CPU: </strong>50\t💾 <strong>Memory %: </strong> 66.666667</li>" +
			"<li><code>APIC 2</code> -> \t💻 <strong>CPU: </strong>40\t💾 <strong>Memory %: </strong> 60.000000</li></ul>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	// Test Unreachable APIC
	t.Run("Error APIC unreachable", func(t *testing.T) {
		wmc := webex.WebexMockClient
		wmc.SetDefaultFunctions()
		amc := apic.ApicMockClient
		amc.SetDefaultFunctions()
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/cpu"}, nil
		}
		b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
		reqB := webex.WebexWebhook{
			Name: "test-bot",
			Data: &webex.WebexWebhookData{
				RoomId: "AbC13",
			},
		}
		amc.GetProcEntityF = func() ([]apic.ApicMoAttributes, error) {
			return []apic.ApicMoAttributes{{}}, errors.New("Generic APIC Error")
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()

		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !. I could not reach the APIC... Are there any issues?"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
}

func TestUtils(t *testing.T) {
	t.Run("cleanCommand - No additional spaces", func(t *testing.T) {

		s := cleanCommand("test-bot", "/ep AA:AA:AA:AA:AA:AA test-bot")
		equals(t, s, "/ep AA:AA:AA:AA:AA:AA")
	})
	t.Run("cleanCommand - Additional spaces & Bot at the end", func(t *testing.T) {

		s := cleanCommand("test-bot", "   /ep   AA:AA:AA:AA:AA:AA   test-bot  ")
		equals(t, s, "/ep AA:AA:AA:AA:AA:AA")
	})
	t.Run("cleanCommand - Additional spaces & Bot at the beginning", func(t *testing.T) {

		s := cleanCommand("test-bot", "test-bot  /ep   AA:AA:AA:AA:AA:AA  ")
		equals(t, s, "/ep AA:AA:AA:AA:AA:AA")
	})
}
