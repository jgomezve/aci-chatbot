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

// Type to define the APIC client configuration option
type Option func(*ApicClient)

// type to define the APIC MO Attributes
type ApicMoAttributes map[string]string

// Struct to store Fabric information. See GetFabricInformation()
type FabricInformation struct {
	Name   string
	Url    string
	Pods   []map[string]string
	Apics  []map[string]string
	Spines []map[string]string
	Leafs  []map[string]string
	Health string
}

// Struct to store Endpoint information. See GetEndpointInformation()
type EndpointInformation struct {
	Mac      string
	Ips      []string
	Location []map[string]string
	Tenant   string
	App      string
	Epg      string
}

// Interface used to mock the HTTP Client
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Apic interface. Implemented by ApicClient and ApicClientMocks
type ApicInterface interface {
	Login() error
	GetIp() string
	GetToken() string
	GetProcEntity() ([]ApicMoAttributes, error)
	SubscribeClassWebSocket(c string) (string, error)
	RefreshSubscriptionWebSocket(id string) error
	GetFabricInformation() (FabricInformation, error)
	GetEndpointInformation(m string) ([]EndpointInformation, error)
	GetFabricNeighbors(nd string) (map[string][]string, error)
	GetLatestFaults(c string) ([]ApicMoAttributes, error)
	GetLatestEvents(c string, usr ...string) ([]ApicMoAttributes, error)
}

// Apic Client struct
type ApicClient struct {
	httpClient HttpClient // It points to an interface. Used for mocking
	usr        string
	pwd        string
	tkn        string
	baseURL    string
}

// Package level variable to define which objects is used as http client (Mock or the standard)
var (
	Client HttpClient
)

// Function executed upon package import. By default the httpClient is the standard http Client form the net/http library
func init() {
	Client = &http.Client{
		Timeout: 3 * time.Second,
	}
}

// The client timeout
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

// Create new APIC client
func NewApicClient(url, usr, pwd string, options ...Option) (*ApicClient, error) {
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

	if err := client.Login(); err != nil {
		return nil, err
	}
	return &client, nil
}

// Get the client URL
func (client *ApicClient) GetIp() string {
	return client.baseURL
}

// Get the current valid token
func (client *ApicClient) GetToken() string {
	return client.tkn
}

// Login to the APIC
func (client *ApicClient) Login() error {

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

// Refresh subscription
func (client *ApicClient) RefreshSubscriptionWebSocket(id string) error {
	var result map[string]interface{}
	req, err := client.makeCall(http.MethodGet, fmt.Sprintf("/api/subscriptionRefresh.json?id=%s", id), nil)

	if err != nil {
		return err
	}
	if err = client.doCall(req, &result); err != nil {
		log.Println("Error: ", err)
		return err
	}
	return nil
}

// Subscribe to class events
func (client *ApicClient) SubscribeClassWebSocket(c string) (string, error) {
	var result map[string]interface{}
	req, err := client.makeCall(http.MethodGet, fmt.Sprintf("/api/class/%s.json?subscription=yes&refresh-timeout=120?query-target=subtree", c), nil)

	if err != nil {
		return "", err
	}
	if err = client.doCall(req, &result); err != nil {
		log.Println("Error: ", err)
		return "", err
	}
	return result["subscriptionId"].(string), nil
}

// Get the latest fabric events.
// Filtering based on username is optional
func (client *ApicClient) GetLatestEvents(c string, usr ...string) ([]ApicMoAttributes, error) {

	q := fmt.Sprintf("order-by=aaaModLR.created|desc&page-size=%s", c)

	for _, u := range usr {
		q += fmt.Sprintf("&query-target-filter=eq(aaaModLR.user,\"%s\")", u)
	}

	events, err := client.reqApicClass(http.MethodGet, "aaaModLR", q)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// Get the latest fabric fault.
// Filtering based on username is optional
func (client *ApicClient) GetLatestFaults(c string) ([]ApicMoAttributes, error) {

	faults, err := client.reqApicClass(http.MethodGet, "faultInst", "order-by=faultInst.lastTransition|desc", fmt.Sprintf("page-size=%s", c))
	if err != nil {
		return nil, err
	}
	return faults, nil
}

// Get the Fabric LLDP and CDP neigh
// Filter based on node id optional
func (client *ApicClient) GetFabricNeighbors(nd string) (map[string][]string, error) {

	cdpN, err := client.reqApicClass(http.MethodGet, "cdpAdjEp")
	if err != nil {
		return nil, err
	}
	lldpN, err := client.reqApicClass(http.MethodGet, "lldpAdjEp")
	if err != nil {
		return nil, err
	}
	neighMap := make(map[string][]string)

	for _, n := range append(cdpN, lldpN...) {
		node := GetRn(n["dn"], "node")
		nodeIface := fmt.Sprintf("%s:%s", node, GetRn(n["dn"], "if"))
		if !stringInSlice(nodeIface, neighMap[n["sysName"]]) && (nd == node || nd == "all") && n["sysName"] != "" {
			neighMap[n["sysName"]] = append(neighMap[n["sysName"]], nodeIface)
		}
	}
	return neighMap, nil
}

// Get information of the fabric
// Number of switches, Pods, Health
func (client *ApicClient) GetFabricInformation() (FabricInformation, error) {

	var info FabricInformation

	// Get the values from the APIC
	banner, err := client.reqApicClass(http.MethodGet, "aaaPreLoginBanner")
	if err != nil {
		return FabricInformation{}, err
	}
	pods, err := client.reqApicClass(http.MethodGet, "fabricPod")
	if err != nil {
		return FabricInformation{}, err
	}
	nodes, err := client.reqApicClass(http.MethodGet, "fabricNode")
	if err != nil {
		return FabricInformation{}, err
	}
	health, err := client.reqApicClass(http.MethodGet, "fabricOverallHealthHist5min", "query-target-filter=and(eq(fabricOverallHealthHist5min.dn,\"topology/HDfabricOverallHealth5min-0\"))")
	if err != nil {
		return FabricInformation{}, err
	}
	//Parse result
	info.Name = banner[0]["guiTextMessage"]
	info.Pods = make([]map[string]string, 0)

	for _, item := range pods {
		info.Pods = append(info.Pods, map[string]string{"id": item["id"], "type": item["podType"]})
	}
	info.Spines = make([]map[string]string, 0)
	info.Leafs = make([]map[string]string, 0)
	info.Apics = make([]map[string]string, 0)
	for _, item := range nodes {
		switch item["role"] {
		case "controller":
			info.Apics = append(info.Apics, map[string]string{"name": item["name"], "version": item["version"]})
		case "leaf":
			info.Leafs = append(info.Leafs, map[string]string{"name": item["name"], "version": item["version"]})
		case "spine":
			info.Spines = append(info.Spines, map[string]string{"name": item["name"], "version": item["version"]})
		}
	}
	info.Health = health[0]["healthAvg"]
	info.Url = client.baseURL
	return info, nil
}

// Get information from an specific enpoint [MAC]
func (client *ApicClient) GetEndpointInformation(m string) ([]EndpointInformation, error) {
	var info []EndpointInformation
	ep, err := client.reqApicClass(http.MethodGet, "fvCEp", fmt.Sprintf("query-target-filter=eq(fvCEp.mac,\"%s\")", m))
	if err != nil {
		return []EndpointInformation{}, err
	}
	for _, itemEp := range ep {
		var ep EndpointInformation
		ep.Mac = itemEp["mac"]
		ep.Tenant = GetRn(itemEp["dn"], "tn")
		ep.App = GetRn(itemEp["dn"], "ap")
		ep.Epg = GetRn(itemEp["dn"], "epg")
		// Only return EPG Endpoints
		if ep.Epg == "" {
			continue
		}
		ips, err := client.getMoChildren("fvCEp", "fvIp", fmt.Sprintf("eq(fvCEp.dn,\"%s\")", itemEp["dn"]))
		if err != nil {
			return []EndpointInformation{}, err
		}
		for _, itempIp := range ips {
			ep.Ips = append(ep.Ips, itempIp["addr"])
		}
		paths, err := client.getMoChildren("fvCEp", "fvRsCEpToPathEp", fmt.Sprintf("eq(fvCEp.dn,\"%s\")", itemEp["dn"]))
		if err != nil {
			return []EndpointInformation{}, err
		}
		for _, itempPath := range paths {
			location := getPath(itempPath["tDn"])
			if location != nil {
				ep.Location = append(ep.Location, location)
			}

		}

		info = append(info, ep)
	}
	return info, nil
}

// Get procEntity class
func (client *ApicClient) GetProcEntity() ([]ApicMoAttributes, error) {
	proc, err := client.reqApicClass(http.MethodGet, "procEntity")
	if err != nil {
		return nil, err
	}
	return proc, nil
}

// Get children Objects from an specific class
func (client *ApicClient) getMoChildren(parent, children, query string) ([]ApicMoAttributes, error) {
	var result map[string]interface{}
	url := fmt.Sprintf("/api/node/class/%s.json?rsp-subtree=children&rsp-subtree-class=%s&query-target-filter=%s", parent, children, query)
	req, err := client.makeCall(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	if err = client.doCall(req, &result); err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	return getApicManagedObjectsChildren(result, parent, children), nil
}

// Generic request for an APIC Class/MO. Only works for the /apic/node/class URI
// Server-side filtering is optional
func (client *ApicClient) reqApicClass(m, c string, filter ...string) ([]ApicMoAttributes, error) {
	var result map[string]interface{}
	url := fmt.Sprintf("/api/node/class/%s.json", c)

	// TODO: How to improve this
	if len(filter) > 0 {
		url += "?"
	}
	for _, f := range filter {
		url += "&" + f
	}
	req, err := client.makeCall(m, url, nil)

	if err != nil {
		return nil, err
	}

	if err = client.doCall(req, &result); err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	return getApicManagedObjects(result, c), nil
}

// Create HTTP Request
func (client *ApicClient) makeCall(m, url string, p io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(m, client.baseURL+url, p)
	if err != nil {
		return nil, errors.New("unable to create a new HTTP request")
	}

	req.Header.Add("Accept", "application/json")
	if url != "/api/aaaLogin.json" {
		req.Header.Set("Cookie", "APIC-cookie="+client.tkn)
	}

	return req, nil
}

// Execute HTTP Request
func (client *ApicClient) doCall(req *http.Request, res interface{}) error {
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error processing this request %s\n API message %s", req.URL, body)
	}

	if err = json.Unmarshal(body, &res); err != nil {
		return err
	}
	return nil
}
