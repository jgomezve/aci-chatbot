package apic

import (
	"aci-chatbot/mocks"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Option func(*ApicClient)

type ApicMoAttributes map[string]string

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ApicClient struct {
	httpClient HttpClient
	usr        string
	pwd        string
	tkn        string
	baseURL    string
}

var (
	Client HttpClient
)

func init() {
	Client = &http.Client{
		Timeout: 3 * time.Second,
	}
}

func SetTimeout(t time.Duration) Option {
	return func(client *ApicClient) {
		switch client.httpClient.(type) {
		case *http.Client:
			client.httpClient.(*http.Client).Timeout = t * time.Second
		case *mocks.MockClient:
			client.httpClient.(*mocks.MockClient).Timeout = t * time.Second
		}
	}
}

func NewApicClient(url, usr, pwd string, options ...Option) (ApicClient, error) {
	client := ApicClient{
		usr:        usr,
		pwd:        pwd,
		httpClient: Client,
		baseURL:    url,
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	for _, opt := range options {
		opt(&client)
	}

	if err := client.login(); err != nil {
		return client, err
	}
	return client, nil
}

func (client *ApicClient) login() error {

	var result map[string]interface{}
	loginPayload := fmt.Sprintf(`{"aaaUser":{"attributes":{"name":"%s","pwd":"%s"}}}`, client.usr, client.pwd)
	req, err := client.makeCall(http.MethodPost, "/api/aaaLogin.json", strings.NewReader(loginPayload))
	if err != nil {
		return err
	}

	if err = client.doCall(req, &result); err != nil {
		log.Println("Error: ", err)
		return err
	}

	r := getApicManagedObjects(result, "aaaLogin")
	client.tkn = r[0]["token"]
	return nil
}

func (client *ApicClient) GetEnpoint(mac string) []ApicMoAttributes {
	var result map[string]interface{}
	url := "/api/node/class/fvCEp.json?query-target-filter=eq(fvCEp.mac,\"" + mac + "\")"
	req, err := client.makeCall(http.MethodGet, url, nil)

	if err != nil {
		return nil
	}

	if err = client.doCall(req, &result); err != nil {
		log.Println("Error: ", err)
		return nil
	}
	return getApicManagedObjects(result, "fvCEp")
}

func (client *ApicClient) GetProcEntity() []ApicMoAttributes {
	var result map[string]interface{}
	req, err := client.makeCall(http.MethodGet, "/api/node/class/procEntity.json", nil)

	if err != nil {
		return nil
	}

	if err = client.doCall(req, &result); err != nil {
		log.Println("Error: ", err)
		return nil
	}
	return getApicManagedObjects(result, "procEntity")
}

func (client *ApicClient) makeCall(m string, url string, p io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(m, client.baseURL+url, p)
	if err != nil {
		return nil, errors.New("unable to create a new HTTP request")
	}

	req.Header.Add("Accept", "application/json")
	// req.Header.Add("Content-Type", "application/json")
	if url != "/api/aaaLogin.json" {
		req.Header.Set("Cookie", "APIC-cookie="+client.tkn)
	}

	return req, nil
}

func (client *ApicClient) doCall(req *http.Request, res interface{}) error {
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return errors.New("unable to send the HTTP request")
	}

	// Why defer ?
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return errors.New("unable to read the response body")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error processing this request %s\n API message %s", req.URL, body)
	}

	if err = json.Unmarshal(body, &res); err != nil {
		return errors.New("unable to read the response body")
	}
	return nil
}
