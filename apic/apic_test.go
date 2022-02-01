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

// Helper functions
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

func notOk(tb testing.TB, err error) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

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
		clt, err := NewApicClient("http://mocking.com", "admin", "admin")
		ok(t, err)
		equals(t, clt.baseURL, "http://mocking.com")
		equals(t, clt.tkn, "eyJhbGciOiJSUzI1NiIsImtpZCI6InJqcmRjazBuNW")
	})

	t.Run("Create Client with Timeout", func(t *testing.T) {
		mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
			r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
			return &http.Response{StatusCode: 200, Body: r}, nil
		}
		clt, err := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
		ok(t, err)
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
		notOk(t, err)
		equals(t, err, errors.New("Generic HTTP Error"))
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
		info, err := clt.GetFabricInformation()
		ok(t, err)
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
		info, err := clt.GetEndpointInformation("00:50:56:96:A0:D3")
		ok(t, err)
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

func TestGetFabricNeighbors(t *testing.T) {
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
	cdp := `{
		"totalCount": "2",
		"imdata": [
			{
				"cdpAdjEp": {
					"attributes": {
						"dn": "topology/pod-1/node-101/sys/cdp/inst/if-[eth1/1]/adj-1",
						"sysName": "SW-1"
					}
				}
			},
			{
				"cdpAdjEp": {
					"attributes": {
						"dn": "topology/pod-1/node-101/sys/cdp/inst/if-[eth1/2]/adj-1",
						"sysName": "SW-2"
					}
				}
			}
		]
	}`
	lldp := `{
		"totalCount": "2",
		"imdata": [
			{
				"lldpAdjEp": {
					"attributes": {
						"dn": "topology/pod-1/node-102/sys/cdp/inst/if-[eth2/1]/adj-1",
						"sysName": "SW-3"
					}
				}
			},
			{
				"lldpAdjEp": {
					"attributes": {
						"dn": "topology/pod-1/node-101/sys/cdp/inst/if-[eth2/2]/adj-1",
						"sysName": "SW-2"
					}
				}
			}
		]
	}`
	mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "aaaLogin") {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(login)))}, nil
		} else if strings.Contains(req.URL.Path, "/api/node/class/cdpAdjEp.json") {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(cdp)))}, nil
		} else if strings.Contains(req.URL.Path, "/api/node/class/lldpAdjEp.json") {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(lldp)))}, nil
		}
		return nil, nil
	}
	clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
	t.Run("All neighbors", func(t *testing.T) {
		neigh, err := clt.GetFabricNeighbors("all")
		ok(t, err)
		equals(t, len(neigh), 3)
		equals(t, neigh["SW-1"], []string{"101:[eth1/1]"})
		equals(t, neigh["SW-2"], []string{"101:[eth1/2]", "101:[eth2/2]"})
		equals(t, neigh["SW-3"], []string{"102:[eth2/1]"})
	})

	t.Run("Node 101", func(t *testing.T) {
		neigh, err := clt.GetFabricNeighbors("101")
		ok(t, err)
		equals(t, len(neigh), 2)
		equals(t, neigh["SW-1"], []string{"101:[eth1/1]"})
		equals(t, neigh["SW-2"], []string{"101:[eth1/2]", "101:[eth2/2]"})
	})
}

func TestGetLatestFaults(t *testing.T) {
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
	faults := `{
		"totalCount": "1",
		"imdata": [
			{
				"faultInst": {
					"attributes": {
						"ack": "no",
						"alert": "no",
						"cause": "resolution-failed",
						"changeSet": "state (Old: formed, New: missing-target)",
						"childAction": "",
						"code": "F1123",
						"created": "2021-10-31T20:15:05.725+01:00",
						"delegated": "no",
						"descr": "Failed to form relation to MO uni/tn-tenant/brc-CON_A of class vzBrCP",
						"dn": "uni/tn-tenant/cif-CON_IFACE/rsif/fault-F1123",
						"domain": "infra",
						"highestSeverity": "warning",
						"lastTransition": "2021-10-31T20:15:05.725+01:00",
						"lc": "raised",
						"occur": "1",
						"origSeverity": "warning",
						"prevSeverity": "warning",
						"rule": "vz-rs-if-resolve-fail",
						"severity": "warning",
						"status": "",
						"subject": "relation-resolution",
						"title": "null",
						"type": "config"
					}
				}
			}
		]
	}`
	mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "aaaLogin") {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(login)))}, nil
		} else if strings.Contains(req.URL.Path, "/api/node/class/faultInst.json") {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(faults)))}, nil
		}
		return nil, nil
	}
	clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
	t.Run("Errorless Call", func(t *testing.T) {
		fault, err := clt.GetLatestFaults("all")
		ok(t, err)
		equals(t, len(fault), 1)
		equals(t, fault[0]["dn"], "uni/tn-tenant/cif-CON_IFACE/rsif/fault-F1123")
		equals(t, fault[0]["severity"], "warning")
	})

}
func TestGetLatestEvents(t *testing.T) {
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
	events := `{
		"totalCount": "3",
		"imdata": [
			{
				"aaaModLR": {
					"attributes": {
						"affected": "uni/tn-myTenant/ap-AP1/epg-EP1",
						"cause": "transition",
						"changeSet": "",
						"childAction": "",
						"clientTag": "",
						"code": "E4211938",
						"created": "2021-10-28T07:32:34.846+01:00",
						"descr": "AEPg EP1 deleted",
						"dn": "subj-[uni/tn-myTenant/ap-AP1/epg-EP1]/mod-4295233655",
						"id": "4295233655",
						"ind": "deletion",
						"modTs": "never",
						"sessionId": "iEesZLm8RRy7B23FqyW3eQ==",
						"severity": "info",
						"status": "",
						"trig": "config",
						"txId": "576460752332323773",
						"user": "user1"
					}
				}
			}
		]
	}`
	mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "aaaLogin") {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(login)))}, nil
		} else if strings.Contains(req.URL.Path, "/api/node/class/aaaModLR.json") {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(events)))}, nil
		}
		return nil, nil
	}
	clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
	t.Run("Get All Events", func(t *testing.T) {
		fault, err := clt.GetLatestEvents("1")
		ok(t, err)
		equals(t, len(fault), 1)
		equals(t, fault[0]["dn"], "subj-[uni/tn-myTenant/ap-AP1/epg-EP1]/mod-4295233655")
		equals(t, fault[0]["user"], "user1")
	})
	t.Run("Get Events from User", func(t *testing.T) {
		fault, err := clt.GetLatestEvents("1", "user1")
		ok(t, err)
		equals(t, len(fault), 1)
		equals(t, fault[0]["dn"], "subj-[uni/tn-myTenant/ap-AP1/epg-EP1]/mod-4295233655")
		equals(t, fault[0]["user"], "user1")
	})

}

func TestSubscribeClassWebSocket(t *testing.T) {
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
	subscription := `{
		"totalCount": "1",
		"subscriptionId": "72079567111979009",
		"imdata": [
			{
				"fvTenant": {
					"attributes": {
						"annotation": "",
						"childAction": "",
						"descr": "",
						"dn": "uni/tn-myTenant",
						"extMngdBy": "",
						"lcOwn": "local",
						"modTs": "2021-02-18T21:55:58.121+01:00",
						"monPolDn": "uni/tn-common/monepg-default",
						"name": "myTenant",
						"nameAlias": "",
						"ownerKey": "",
						"ownerTag": "",
						"status": "",
						"uid": "9152",
						"userdom": ":all:mgmt:common:"
					}
				}
			}
		]
	}`
	mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "aaaLogin") {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(login)))}, nil
		} else if strings.Contains(req.URL.Path, "/api/class/fvTenant") {
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(subscription)))}, nil
		}
		return nil, nil
	}
	clt, _ := NewApicClient("http://mocking.com", "admin", "admin", SetTimeout(8))
	t.Run("Get All Events", func(t *testing.T) {
		subId, err := clt.SubscribeClassWebSocket("fvTenant")
		ok(t, err)
		equals(t, subId, "72079567111979009")
	})

}
