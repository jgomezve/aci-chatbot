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

type FabricInformation struct {
	Name   string
	Url    string
	Pods   []map[string]string
	Apics  []map[string]string
	Spines []map[string]string
	Leafs  []map[string]string
	Health string
}

type EndpointInformation struct {
	Mac      string
	Ips      []string
	Location []map[string]string
	Tenant   string
	App      string
	Epg      string
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ApicInterface interface {
	GetProcEntity() ([]ApicMoAttributes, error)
	GetFabricInformation() (FabricInformation, error)
	GetEndpointInformation(m string) ([]EndpointInformation, error)
	GetFabricNeighbors(nd string) (map[string][]string, error)
	GetLastestFaults(c string) ([]ApicMoAttributes, error)
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

// TODO: Use a generic version of _getApicClass_ that uses all the HTTP Verbs
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

func (client *ApicClient) GetLastestFaults(c string) ([]ApicMoAttributes, error) {

	faults, err := client.getApicClass("faultInst", "order-by=faultInst.lastTransition|desc", fmt.Sprintf("page-size=%s", c))
	if err != nil {
		return nil, err
	}
	return faults, nil
}

func (client *ApicClient) GetFabricNeighbors(nd string) (map[string][]string, error) {

	cdpN, err := client.getApicClass("cdpAdjEp")
	if err != nil {
		return nil, err
	}
	lldpN, err := client.getApicClass("lldpAdjEp")
	if err != nil {
		return nil, err
	}
	neighMap := make(map[string][]string)

	for _, n := range append(cdpN, lldpN...) {
		node := GetRn(n["dn"], "node")
		nodeIface := fmt.Sprintf("%s:%s", node, GetRn(n["dn"], "if"))
		if !stringInSlice(nodeIface, neighMap[n["sysName"]]) && (nd == node || nd == "") && n["sysName"] != "" {
			neighMap[n["sysName"]] = append(neighMap[n["sysName"]], nodeIface)
		}
	}
	return neighMap, nil
}

func (client *ApicClient) GetFabricInformation() (FabricInformation, error) {

	var info FabricInformation

	// Get the values from the APIC
	banner, err := client.getApicClass("aaaPreLoginBanner")
	if err != nil {
		return FabricInformation{}, err
	}
	pods, err := client.getApicClass("fabricPod")
	if err != nil {
		return FabricInformation{}, err
	}
	nodes, err := client.getApicClass("fabricNode")
	if err != nil {
		return FabricInformation{}, err
	}
	health, err := client.getApicClass("fabricOverallHealthHist5min", "query-target-filter=and(eq(fabricOverallHealthHist5min.dn,\"topology/HDfabricOverallHealth5min-0\"))")
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

func (client *ApicClient) GetEndpointInformation(m string) ([]EndpointInformation, error) {
	var info []EndpointInformation
	ep, err := client.getApicClass("fvCEp", fmt.Sprintf("query-target-filter=eq(fvCEp.mac,\"%s\")", m))
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

func (client *ApicClient) GetProcEntity() ([]ApicMoAttributes, error) {
	proc, err := client.getApicClass("procEntity")
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (client *ApicClient) getMoChildren(parent string, children string, query string) ([]ApicMoAttributes, error) {
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

func (client *ApicClient) getApicClass(n string, filter ...string) ([]ApicMoAttributes, error) {
	var result map[string]interface{}
	url := fmt.Sprintf("/api/node/class/%s.json", n)

	// TODO: How to improve this
	if len(filter) > 0 {
		url += "?"
	}
	for _, f := range filter {
		url += "&" + f
	}
	req, err := client.makeCall(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	if err = client.doCall(req, &result); err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	return getApicManagedObjects(result, n), nil
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
		// TODO: Check error message
		return errors.New("unable to read the response body")
	}
	return nil
}
