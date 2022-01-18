package apic

import (
	"aci-chatbot/mocks"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Helper function
func equals(tb testing.TB, act, exp interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func TestApicClientCreation(t *testing.T) {

	Client = &mocks.MockClient{}
	json := `{
		"totalCount": "1",
		"imdata": [
			{
				"aaaLogin": {
					"attributes": {
						"token": "eyJhbGciOiJSUzI1NiIsImtpZCI6InJqcmRjazBuNW"
					}
				}
			}
		]
	}`
	t.Run("Create Raw Client", func(t *testing.T) {

		mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
			r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
			return &http.Response{StatusCode: 200, Body: r}, nil
		}
		clt, _ := NewApicClient("http://mocking.com", "admin", "admin")
		equals(t, clt.baseURL, "http://mocking.com")
		equals(t, clt.tkn, "eyJhbGciOiJSUzI1NiIsImtpZCI6InJqcmRjazBuNW")
	})

	t.Run("Create Client with Timeout", func(t *testing.T) {
		mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
			r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
			return &http.Response{StatusCode: 200, Body: r}, nil
		}
		clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
		equals(t, clt.baseURL, "http://mocking.com")
		equals(t, clt.tkn, "eyJhbGciOiJSUzI1NiIsImtpZCI6InJqcmRjazBuNW")
		equals(t, clt.httpClient.(*mocks.MockClient).Timeout, 8*time.Second)
	})
}

func TestGetProcEntity(t *testing.T) {

	Client = &mocks.MockClient{}
	login := `{
		"totalCount": "1",
		"imdata": [
			{
				"aaaLogin": {
					"attributes": {
						"token": "eyJhbGciOiJSUzI1NiIsImtpZCI6InJqcmRjazBuNW"
					}
				}
			}
		]
	}`
	proc := `{
		"totalCount": "2",
		"imdata": [
			{
				"procEntity": {
					"attributes": {
						"cpuPct": "26",
						"dn": "topology/pod-1/node-1/sys/proc",
						"maxMemAlloc": "131427900",
						"memFree": "76078044"
					}
				}
			},
			{
				"procEntity": {
					"attributes": {
						"cpuPct": "11",
						"dn": "topology/pod-1/node-2/sys/proc",
						"maxMemAlloc": "196438552",
						"memFree": "162262988"
					}
				}
			}
		]
	}`
	t.Run("Sucessfull Call", func(t *testing.T) {
		mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
			loginR := ioutil.NopCloser(bytes.NewReader([]byte(login)))
			procR := ioutil.NopCloser(bytes.NewReader([]byte(proc)))
			if strings.Contains(req.URL.Path, "aaaLogin") {
				return &http.Response{StatusCode: 200, Body: loginR}, nil
			} else if strings.Contains(req.URL.Path, "procEntity") {
				return &http.Response{StatusCode: 200, Body: procR}, nil
			}
			return nil, nil
		}
		clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
		procs, _ := clt.GetProcEntity()
		equals(t, procs[0]["dn"], "topology/pod-1/node-1/sys/proc")
		equals(t, procs[0]["cpuPct"], "26")
		equals(t, procs[0]["maxMemAlloc"], "131427900")
		equals(t, procs[0]["memFree"], "76078044")

		equals(t, procs[1]["dn"], "topology/pod-1/node-2/sys/proc")
		equals(t, procs[1]["cpuPct"], "11")
		equals(t, procs[1]["maxMemAlloc"], "196438552")
		equals(t, procs[1]["memFree"], "162262988")

	})

	t.Run("API Error", func(t *testing.T) {
		mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
			loginR := ioutil.NopCloser(bytes.NewReader([]byte(login)))
			if strings.Contains(req.URL.Path, "aaaLogin") {
				return &http.Response{StatusCode: 200, Body: loginR}, nil
			} else {
				return &http.Response{StatusCode: 200, Body: nil}, errors.New("Generic HTTP Error")
			}
		}
		clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
		_, err := clt.GetProcEntity()
		equals(t, err, errors.New("unable to send the HTTP request"))
	})
}

func TestGetFabricInformation(t *testing.T) {

	Client = &mocks.MockClient{}
	login := `{
		"totalCount": "1",
		"imdata": [
			{
				"aaaLogin": {
					"attributes": {
						"token": "eyJhbGciOiJSUzI1NiIsImtpZCI6InJqcmRjazBuNW"
					}
				}
			}
		]
	}`
	banner := `{
		"totalCount": "1",
		"imdata": [
			{
				"aaaPreLoginBanner": {
					"attributes": {
						"guiTextMessage": "CX Fabric"
					}
				}
			}
		]
	}`
	pods := `
	{
		"totalCount": "1",
		"imdata": [
			{
				"fabricPod": {
					"attributes": {
						"id": "1",
						"podType": "physical"
					}
				}
			}
		]
	}`
	nodes := `{
		"totalCount": "3",
		"imdata": [
			{
				"fabricNode": {
					"attributes": {
						"name": "Node-1002",
						"role": "leaf",
						"version": "n9000-15.2(2f)"
					}
				}
			},
			{
				"fabricNode": {
					"attributes": {
						"name": "apic1",
						"role": "controller",
						"version": "5.2(2f)"
					}
				}
			},
			{
				"fabricNode": {
					"attributes": {
						"name": "Node-101",
						"role": "spine",
						"version": "n9000-15.2(2f)"
					}
				}
			}
		]
	}`
	health := `{
		"totalCount": "1",
		"imdata": [
			{
				"fabricOverallHealthHist5min": {
					"attributes": {
						"healthAvg": "95"
					}
				}
			}
		]
	}`
	t.Run("Successfull Call", func(t *testing.T) {
		mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "aaaLogin") {
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(login)))}, nil
			} else if strings.Contains(req.URL.Path, "aaaPreLoginBanner") {
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(banner)))}, nil
			} else if strings.Contains(req.URL.Path, "fabricPod") {
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(pods)))}, nil
			} else if strings.Contains(req.URL.Path, "fabricNode") {
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(nodes)))}, nil
			} else if strings.Contains(req.URL.Path, "fabricOverallHealthHist5min") {
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(health)))}, nil
			}
			return nil, nil
		}
		clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
		info, _ := clt.GetFabricInformation()
		equals(t, info.Name, "CX Fabric")
		equals(t, info.Health, "95")
		equals(t, info.Apics[0]["name"], "apic1")
		equals(t, info.Apics[0]["version"], "5.2(2f)")
		equals(t, info.Spines[0]["name"], "Node-101")
		equals(t, info.Spines[0]["version"], "n9000-15.2(2f)")
		equals(t, info.Leafs[0]["name"], "Node-1002")
		equals(t, info.Leafs[0]["version"], "n9000-15.2(2f)")
		equals(t, info.Pods[0]["id"], "1")
		equals(t, info.Pods[0]["type"], "physical")
		equals(t, info.Url, "http://mocking.com")
		equals(t, len(info.Pods), 1)
		equals(t, len(info.Apics), 1)
		equals(t, len(info.Spines), 1)
		equals(t, len(info.Leafs), 1)
	})
}

func TestGetEndpointInformation(t *testing.T) {

	Client = &mocks.MockClient{}
	login := `{
		"totalCount": "1",
		"imdata": [
			{
				"aaaLogin": {
					"attributes": {
						"token": "eyJhbGciOiJSUzI1NiIsImtpZCI6InJqcmRjazBuNW"
					}
				}
			}
		]
	}`
	ep := `{
		"totalCount": "1",
		"imdata": [
			{
				"fvCEp": {
					"attributes": {
						"dn": "uni/tn-test_tenant/ap-Test-AP/epg-test_epg/cep-00:50:56:96:A0:D3",
						"mac": "00:50:56:96:A0:D3"
					}
				}
			}
		]
	}`
	ips := `{
		"totalCount": "1",
		"imdata": [
			{
				"fvCEp": {
					"attributes": {
						"dn": "uni/tn-test_tenant/ap-Test-AP/epg-test_epg/cep-00:50:56:96:A0:D3",
						"mac": "00:50:56:96:4B:59"
					},
					"children": [
						{
							"fvIp": {
								"attributes": {
									"addr": "172.20.206.132"
								}
							}
						},
						{
							"fvIp": {
								"attributes": {
									"addr": "172.20.206.133"
								}
							}
						}
					]
				}
			}
		]
	}`
	paths := `{
		"totalCount": "1",
		"imdata": [
			{
				"fvCEp": {
					"attributes": {
						"dn": "uni/tn-test_tenant/ap-Test-AP/epg-test_epg/cep-00:50:56:96:A0:D3",
						"mac": "00:50:56:96:4B:59"
					},
					"children": [
						{
							"fvRsCEpToPathEp": {
								"attributes": {
									"tDn": "topology/pod-2/protpaths-1201-1202/pathep-[VPC_TEST_IPG]"
								}
							}
						},
						{
							"fvRsCEpToPathEp": {
								"attributes": {
									"tDn": "topology/pod-2/paths-1201/pathep-[PC_TEST_IPG]"
								}
							}
						},
						{
							"fvRsCEpToPathEp": {
								"attributes": {
									"tDn": "topology/pod-2/paths-1202/pathep-[eth1/10]"
								}
							}
						},
						{
							"fvRsCEpToPathEp": {
								"attributes": {
									"tDn": "topology/pod-2/paths/pathep-[tunnel123]"
								}
							}
						}
					]
				}
			}
		]
	}`
	t.Run("Successfull Call", func(t *testing.T) {
		mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "aaaLogin") {
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(login)))}, nil
			} else if strings.Contains(req.URL.Path, "/api/node/class/fvCEp.json") && strings.Contains(req.URL.RawQuery, "query-target-filter=eq(fvCEp.mac") {
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(ep)))}, nil
			} else if strings.Contains(req.URL.Path, "/api/node/class/fvCEp.json") && strings.Contains(req.URL.RawQuery, "rsp-subtree-class=fvIp") {
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(ips)))}, nil
			} else if strings.Contains(req.URL.Path, "/api/node/class/fvCEp.json") && strings.Contains(req.URL.RawQuery, "rsp-subtree-class=fvRsCEpToPathEp") {
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(paths)))}, nil
			}
			return nil, nil
		}
		clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
		info, _ := clt.GetEndpointInformation("00:50:56:96:A0:D3")
		equals(t, info[0].Mac, "00:50:56:96:A0:D3")
		equals(t, info[0].Tenant, "test_tenant")
		equals(t, info[0].App, "Test-AP")
		equals(t, info[0].Epg, "test_epg")
		equals(t, len(info[0].Ips), 2)
		equals(t, info[0].Ips, []string{"172.20.206.132", "172.20.206.133"})
		equals(t, len(info[0].Location), 3)
		equals(t, info[0].Location[0], map[string]string{"nodes": "1201-1202", "pod": "2", "port": "[VPC_TEST_IPG]", "type": "vPC"})
		equals(t, info[0].Location[1], map[string]string{"nodes": "1201", "pod": "2", "port": "[PC_TEST_IPG]", "type": "PC"})
		equals(t, info[0].Location[2], map[string]string{"nodes": "1202", "pod": "2", "port": "[eth1/10]", "type": "Access"})
	})
}
