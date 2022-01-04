package apic

import (
	"chatbot/mocks"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// ---------- JSON Structs
type ApicLoginReply struct {
	TotalCount string               `json:"totalCount"`
	Imdata     []ApicAaaLoginImData `json:"imdata"`
}
type ApicProcEntityReply struct {
	TotalCount string                       `json:"totalCount"`
	Imdata     []ApicProcEnitityLoginImData `json:"imdata"`
}

type ApicProcEnitityLoginImData struct {
	ProcEntity ApicProcEntity `json:"procEntity"`
}
type ApicAaaLoginImData struct {
	AaaLogin ApicAaaLogin `json:"aaaLogin"`
}
type ApicAaaLogin struct {
	Attributes ApicLoginAttributes `json:"attributes"`
	Children   interface{}         `json:"children"`
}
type ApicProcEntity struct {
	Attributes ApicProcEntityAttributes `json:"attributes"`
	Children   interface{}              `json:"children"`
}

type ApicProcEntityAttributes struct {
	CpuPct  string `json:"cpuPct"`
	MemFree string `json:"memFree"`
}

type ApicLoginAttributes struct {
	Token string `json:"token"`
}

// ----------
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
type Option func(*ApicClient)

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

// --------- Client

var (
	Client HttpClient
)

func init() {
	Client = &http.Client{
		Timeout: 3 * time.Second,
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

	var result ApicLoginReply
	loginPayload := fmt.Sprintf(`{"aaaUser":{"attributes":{"name":"%s","pwd":"%s"}}}`, client.usr, client.pwd)
	status, content, err := client.makeCall(http.MethodPost, "/api/aaaLogin.json", strings.NewReader(loginPayload))
	if err != nil {
		return err
	}
	if status != 200 {
		return errors.New("authentication failed")

	}

	err = json.Unmarshal(content, &result)
	if err != nil {
		return err
	}
	client.tkn = result.Imdata[0].AaaLogin.Attributes.Token
	return nil

}

func (client *ApicClient) GetProcEntity() interface{} {
	var result ApicProcEntityReply
	status, content, err := client.makeCall(http.MethodGet, "/api/node/class/procEntity.json", nil)

	if err != nil {
		return nil
	}
	if status != 200 {
		return nil

	}

	err = json.Unmarshal(content, &result)
	if err != nil {
		return nil
	}

	return result.Imdata[0].ProcEntity

}

func (client *ApicClient) makeCall(m string, url string, p io.Reader) (int, []byte, error) {
	req, err := http.NewRequest(m, client.baseURL+url, p)
	if err != nil {
		return 0, nil, errors.New("unable to create a new HTTP request")
	}

	req.Header.Add("Accept", "application/json")
	// req.Header.Add("Content-Type", "application/json")
	if url != "/api/aaaLogin.json" {
		req.Header.Set("Cookie", "APIC-cookie="+client.tkn)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return 0, nil, errors.New("unable to send the HTTP request")
	}

	// Why defer ?
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return 0, nil, errors.New("unable to read the response body")
	}
	return resp.StatusCode, body, nil

}
