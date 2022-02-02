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
		wmc.DeleteWebhookF = func(id string) error {
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
	wmc := webex.WebexMockClient
	wmc.SetDefaultFunctions()
	b, _ := NewBot(&wmc, nil, "http://test_bot.com")
	// HTTP request to the /test URI without errors
	t.Run("Errorless Request", func(t *testing.T) {

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

func TestWebHookHanlderInfoCommand(t *testing.T) {
	wmc := webex.WebexMockClient
	wmc.SetDefaultFunctions()
	amc := apic.ApicMockClient
	amc.SetDefaultFunctions()
	wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
		return webex.WebexMessage{Text: "/info"}, nil
	}
	b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
	reqB := webex.WebexWebhook{
		Name: "test-bot",
		Data: &webex.WebexWebhookData{
			RoomId: "AbC13",
		},
	}

	t.Run("Errorless /info command", func(t *testing.T) {
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n\n" +
			"This is the general information of the Fabric <code>Test Fabric</code> (https://test-apic.com): \n\n" +
			"<ul><li>Current Health Score: <strong>95</strong></li>" +
			"<li><strong>APIC Controllers</strong><ul>" +
			"<li>APIC1 (<strong>5.2(3e)</strong>)</li></ul>" +
			"<li><strong>Pods</strong><ul>" +
			"<li>Pod1 <em>physical</em></li>" +
			"<li>Pod2 <em>physical</em></li></ul>" +
			"<li></strong>Switches</strong><ul>" +
			"<li># of Spines : <strong>1</strong></li>" +
			"<li># of Leafs : <strong>2</strong></li></ul></ul></li>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("Error APIC unreachable", func(t *testing.T) {
		amc.GetFabricInformationF = func() (apic.FabricInformation, error) {
			return apic.FabricInformation{}, errors.New("Generic APIC Error")
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

func TestWebHookHanlderEndpointCommand(t *testing.T) {
	wmc := webex.WebexMockClient
	wmc.SetDefaultFunctions()
	amc := apic.ApicMockClient
	amc.SetDefaultFunctions()
	wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
		return webex.WebexMessage{Text: "/ep AA:AA:BB:BB:CC:CC"}, nil
	}
	b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
	reqB := webex.WebexWebhook{
		Name: "test-bot",
		Data: &webex.WebexWebhookData{
			RoomId: "AbC13",
		},
	}

	t.Run("Errorless /ep command", func(t *testing.T) {
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n\n" +
			"This is the information for the Endpoint <code>AA:AA:BB:BB:CC:CC</code>" +
			"<ul><li><strong>Tenant</strong>: myTenant</li>" +
			"<li><strong>Application Profile</strong>: myApp</li>" +
			"<li><strong>EPG</strong>: myEPG</li>" +
			"<li><strong>Location 1</strong>: </li><ul>" +
			"<li><strong>Pod</strong>: 1  <strong>Node</strong>: 1201-1202  <strong>Type</strong>: vPC  <strong>Port</strong>: [FI_VPC_IPG]</li></ul" +
			"><li><strong>IPs</strong>: </li><ul>" +
			"<li><strong>IP</strong>: 192.168.1.1</li></ul></ul>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("Error APIC unreachable", func(t *testing.T) {
		amc.GetEndpointInformationF = func(m string) ([]apic.EndpointInformation, error) {
			return []apic.EndpointInformation{}, errors.New("Generic APIC Error")
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

func TestWebHookHanlderneighCommand(t *testing.T) {
	wmc := webex.WebexMockClient
	wmc.SetDefaultFunctions()
	amc := apic.ApicMockClient
	amc.SetDefaultFunctions()
	wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
		return webex.WebexMessage{Text: "/neigh"}, nil
	}
	b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
	reqB := webex.WebexWebhook{
		Name: "test-bot",
		Data: &webex.WebexWebhookData{
			RoomId: "AbC13",
		},
	}

	t.Run("Fabric Neighbors", func(t *testing.T) {
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n\n" +
			"This is the Topology information of the Fabric : \n\n" +
			"<ul><li><strong>SW1</strong>:\t101:[eth1/1]   102:[eth1/2]   </li>" +
			"<li><strong>SW2</strong>:\t101:[eth1/3]   103:[eth1/4]   </li>" +
			"<li><strong>SW3</strong>:\t102:[eth1/5]   103:[eth1/6]   </li></ul>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("Node neighbors", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/neigh 103"}, nil
		}
		amc.GetFabricNeighborsF = func(nd string) (map[string][]string, error) {
			return map[string][]string{"SW2": {"103:[eth1/4]"}, "SW3": {"103:[eth1/6]"}}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n\n" +
			"These are the Neighbors of the Node <code>103</code>: \n\n" +
			"<ul><li><strong>SW2</strong>:\t103:[eth1/4]   </li>" +
			"<li><strong>SW3</strong>:\t103:[eth1/6]   </li></ul>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("Fabric neighbors - No Neighbors", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/neigh"}, nil
		}
		amc.GetFabricNeighborsF = func(nd string) (map[string][]string, error) {
			return map[string][]string{}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n Sorry.. I could not discover the Topology of the Fabric"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("Node neighbors - Invalid Node", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/neigh 999"}, nil
		}
		amc.GetFabricNeighborsF = func(nd string) (map[string][]string, error) {
			return map[string][]string{}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n It seems there are no Neighbors for Node <code>999</code>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("APIC Unavailable", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/neigh"}, nil
		}
		amc.GetFabricNeighborsF = func(nd string) (map[string][]string, error) {
			return map[string][]string{}, errors.New("Generic APIC Error")
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

func TestWebHookHanlderFaultCommand(t *testing.T) {
	wmc := webex.WebexMockClient
	wmc.SetDefaultFunctions()
	amc := apic.ApicMockClient
	amc.SetDefaultFunctions()
	wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
		return webex.WebexMessage{Text: "/faults"}, nil
	}
	b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
	reqB := webex.WebexWebhook{
		Name: "test-bot",
		Data: &webex.WebexWebhookData{
			RoomId: "AbC13",
		},
	}

	t.Run("Errorless /fault command", func(t *testing.T) {
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n\n" +
			"These are the latest 10 faults in the the Fabric : \n\n" +
			"<ul><li><strong>F1451</strong> - <em>topology/pod-1/node-202/sys/ch/psuslot-1/psu/fault-F1451</em>" +
			"<ul><li>Power supply shutdown. (serial number ABCDEF)</li>" +
			"<li><strong>Severity</strong>: minor ⚠️</li>" +
			"<li><strong>Current Lyfecycle</strong>: raised ❌</li>" +
			"<li><strong>Type</strong>: environmental</li>" +
			"<li><strong>Created</strong>: 2021-09-07T13:20:13.645+01:00</li></ul></ul>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("Error APIC unreachable", func(t *testing.T) {
		amc.GetLatestFaultsF = func(c string) ([]apic.ApicMoAttributes, error) {
			return []apic.ApicMoAttributes{}, errors.New("Generic APIC Error")
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

func TestWebHookHanlderEventCommand(t *testing.T) {
	wmc := webex.WebexMockClient
	wmc.SetDefaultFunctions()
	amc := apic.ApicMockClient
	amc.SetDefaultFunctions()
	wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
		return webex.WebexMessage{Text: "/events"}, nil
	}
	b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
	reqB := webex.WebexWebhook{
		Name: "test-bot",
		Data: &webex.WebexWebhookData{
			RoomId: "AbC13",
		},
	}

	t.Run("All Users Events", func(t *testing.T) {
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n\n" +
			"These are the latest 10 events in the the Fabric : \n\n" +
			"<ul><li><strong>E4218210</strong> - <em>uni/uipageusage/pagecount-AllTenants</em>" +
			"<ul><li>PageCount AllTenants modified</li>" +
			"<li><strong>User</strong>: user1</li>" +
			"<li><strong>Type</strong>: modification 🔄</li>" +
			"<li><strong>Created</strong>: 2021-09-07T13:20:13.645+01:00</li></ul></ul>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("No Events Returned", func(t *testing.T) {
		amc.GetLatestEventsF = func(c string, usr ...string) ([]apic.ApicMoAttributes, error) {
			return []apic.ApicMoAttributes{}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !. There are no events"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("Specific User Events", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/events usera"}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !. There are no events"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
	t.Run("Error APIC unreachable", func(t *testing.T) {
		amc.GetLatestEventsF = func(c string, usr ...string) ([]apic.ApicMoAttributes, error) {
			return []apic.ApicMoAttributes{}, errors.New("Generic APIC Error")
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

// TODO Mock WebSocket
func TestWebHookHanlderWebsocketCommand(t *testing.T) {
	wmc := webex.WebexMockClient
	wmc.SetDefaultFunctions()
	amc := apic.ApicMockClient
	amc.SetDefaultFunctions()
	b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
	reqB := webex.WebexWebhook{
		Name: "test-bot",
		Data: &webex.WebexWebhookData{
			RoomId: "AbC13",
		},
	}

	t.Run("/websocket - List Subscription - No Subscription", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/websocket list"}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n You are no subscribed to any class"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})

	t.Run("/websocket - Subscribe to Class", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/websocket fvTenant"}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n Websocket subscription to MO/Class <code>fvTenant</code> configured 🔧 !"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})

	t.Run("/websocket - Subscribe to Class Again", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/websocket fvTenant"}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n You are already subscribed to MO/Class <code>fvTenant</code>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})

	t.Run("/websocket - List Subscription", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/websocket list"}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n Here the list of subcribed classes:\n <ul><li><code>fvTenant</code></li></ul>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})

	t.Run("/websocket - Remove Suscription", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/websocket fvTenant rm"}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n Websocket subscription to MO/Class <code>fvTenant</code> deleted 🔧 !"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})

	t.Run("/websocket - Remove Invalid Suscription", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/websocket fvTenant rm"}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 !\n\n You are not subscribed to MO/Class <code>fvTenant</code>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
}

func TestWebHookHanlderHelpCommand(t *testing.T) {
	wmc := webex.WebexMockClient
	wmc.SetDefaultFunctions()
	amc := apic.ApicMockClient
	amc.SetDefaultFunctions()
	wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
		return webex.WebexMessage{Text: "/help"}, nil
	}
	b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
	reqB := webex.WebexWebhook{
		Name: "test-bot",
		Data: &webex.WebexWebhookData{
			RoomId: "AbC13",
		},
	}

	t.Run("Errorless /help command", func(t *testing.T) {
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hello , How can I help you?\n\n" +
			"<ul><li><code>/cpu</code>\t->\tGet APIC CPU Information 💾</li>" +
			"<li><code>/ep</code>\t->\tGet APIC Endpoint Information 💻. Usage <code>/ep [ep_mac] </code></li>" +
			"<li><code>/events</code>\t->\tGet Fabric latest events ❎.   Usage <code>/events [user:opt] [count(1-10):opt] </code></li>" +
			"<li><code>/faults</code>\t->\tGet Fabric latest faults ⚠️. Usage <code>/faults [count(1-10):opt] </code></li>" +
			"<li><code>/help</code>\t->\tChatbot Help ❔</li><li><code>/info</code>\t->\tGet Fabric Information ℹ️</li>" +
			"<li><code>/neigh</code>\t->\tGet Fabric Topology Information 🔢. Usage <code>/neigh [node_id] </code></li>" +
			"<li><code>/websocket</code>\t->\tSubscribe to Fabric events 📩</li><ul>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})
}

func TestWebHookHanlderInvalidCommand(t *testing.T) {
	wmc := webex.WebexMockClient
	wmc.SetDefaultFunctions()
	amc := apic.ApicMockClient
	amc.SetDefaultFunctions()
	wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
		return webex.WebexMessage{Text: "a wrong text"}, nil
	}
	b, _ := NewBot(&wmc, &amc, "http://test_bot.com")
	reqB := webex.WebexWebhook{
		Name: "test-bot",
		Data: &webex.WebexWebhookData{
			RoomId: "AbC13",
		},
	}

	t.Run("Completely wrong command", func(t *testing.T) {
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hello , How can I help you?\n\n" +
			"<ul><li><code>/cpu</code>\t->\tGet APIC CPU Information 💾</li>" +
			"<li><code>/ep</code>\t->\tGet APIC Endpoint Information 💻. Usage <code>/ep [ep_mac] </code></li>" +
			"<li><code>/events</code>\t->\tGet Fabric latest events ❎.   Usage <code>/events [user:opt] [count(1-10):opt] </code></li>" +
			"<li><code>/faults</code>\t->\tGet Fabric latest faults ⚠️. Usage <code>/faults [count(1-10):opt] </code></li>" +
			"<li><code>/help</code>\t->\tChatbot Help ❔</li><li><code>/info</code>\t->\tGet Fabric Information ℹ️</li>" +
			"<li><code>/neigh</code>\t->\tGet Fabric Topology Information 🔢. Usage <code>/neigh [node_id] </code></li>" +
			"<li><code>/websocket</code>\t->\tSubscribe to Fabric events 📩</li><ul>"
		equals(t, wmc.LastMsgSent, expectedMessage)
	})

	t.Run("Incorrect Syntax", func(t *testing.T) {
		wmc.GetMessageByIdF = func(id string) (webex.WebexMessage, error) {
			return webex.WebexMessage{Text: "/neigh abc"}, nil
		}
		jp, _ := json.Marshal(reqB)
		request, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jp))
		response := httptest.NewRecorder()
		b.router.ServeHTTP(response, request)
		equals(t, response.Code, http.StatusOK)
		expectedMessage := "Hi  🤖 \n I could not fully understand the input\n" +
			" Please check the usage of the <code>/neigh</code> command:\n " +
			"<ul><li>Get Fabric Topology Information 🔢. Usage <code>/neigh [node_id] </code></ul></li>\n"
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
