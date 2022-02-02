//go:build !apic
// +build !apic

package apic

type ApicClientMocks struct {
	GetProcEntityF          func() ([]ApicMoAttributes, error)
	GetFabricInformationF   func() (FabricInformation, error)
	GetEndpointInformationF func(m string) ([]EndpointInformation, error)
	GetFabricNeighborsF     func(nd string) (map[string][]string, error)
	GetLatestFaultsF        func(c string) ([]ApicMoAttributes, error)
	GetLatestEventsF        func(c string, usr ...string) ([]ApicMoAttributes, error)
}

// TODO: check if this approach is valid
var (
	ApicMockClient ApicClientMocks
)

// Mock functions default values
func (ac *ApicClientMocks) SetDefaultFunctions() {
	ac.GetProcEntityF = func() ([]ApicMoAttributes, error) {
		procs := []ApicMoAttributes{}
		procs = append(procs, ApicMoAttributes{"dn": "topology/pod-1/node-1/sys/proc", "cpuPct": "50", "memFree": "4000", "maxMemAlloc": "6000"})
		procs = append(procs, ApicMoAttributes{"dn": "topology/pod-1/node-2/sys/proc", "cpuPct": "40", "memFree": "60", "maxMemAlloc": "100"})
		return procs, nil
	}

	ac.GetFabricInformationF = func() (FabricInformation, error) {
		return FabricInformation{
			Name:   "Test Fabric",
			Url:    "https://test-apic.com",
			Pods:   []map[string]string{{"id": "1", "type": "physical"}, {"id": "2", "type": "physical"}},
			Apics:  []map[string]string{{"name": "APIC1", "version": "5.2(3e)"}},
			Spines: []map[string]string{{"name": "SPINE1", "version": "15.2(3e)"}},
			Leafs:  []map[string]string{{"name": "LEAF1", "version": "15.2(3e)"}, {"name": "LEAF2", "version": "15.2(3e)"}},
			Health: "95"}, nil
	}

	ac.GetEndpointInformationF = func(m string) ([]EndpointInformation, error) {
		return []EndpointInformation{{
			Mac:      "AA:AA:AA:BB:BB:CC",
			Ips:      []string{"192.168.1.1"},
			Tenant:   "myTenant",
			Location: []map[string]string{{"pod": "1", "type": "vPC", "nodes": "1201-1202", "port": "[FI_VPC_IPG]"}},
			App:      "myApp",
			Epg:      "myEPG",
		}}, nil
	}

	ac.GetFabricNeighborsF = func(nd string) (map[string][]string, error) {
		return map[string][]string{"SW1": {"101:[eth1/1]", "102:[eth1/2]"}, "SW2": {"101:[eth1/3]", "103:[eth1/4]"}, "SW3": {"102:[eth1/5]", "103:[eth1/6]"}}, nil
	}

	ac.GetLatestFaultsF = func(c string) ([]ApicMoAttributes, error) {
		return []ApicMoAttributes{
			{"code": "F1451",
				"dn":       "topology/pod-1/node-202/sys/ch/psuslot-1/psu/fault-F1451",
				"descr":    "Power supply shutdown. (serial number ABCDEF)",
				"severity": "minor",
				"lc":       "raised",
				"type":     "environmental",
				"created":  "2021-09-07T13:20:13.645+01:00",
			}}, nil
	}

	ac.GetLatestEventsF = func(c string, usr ...string) ([]ApicMoAttributes, error) {
		return []ApicMoAttributes{
			{"code": "E4218210",
				"affected": "uni/uipageusage/pagecount-AllTenants",
				"descr":    "PageCount AllTenants modified",
				"user":     "user1",
				"ind":      "modification",
				"created":  "2021-09-07T13:20:13.645+01:00",
			}}, nil
	}
}

func (ac *ApicClientMocks) GetProcEntity() ([]ApicMoAttributes, error) {
	return ac.GetProcEntityF()
}

func (ac *ApicClientMocks) GetFabricInformation() (FabricInformation, error) {
	return ac.GetFabricInformationF()
}

func (ac *ApicClientMocks) GetEndpointInformation(m string) ([]EndpointInformation, error) {
	return ac.GetEndpointInformationF(m)
}

func (ac *ApicClientMocks) GetFabricNeighbors(nd string) (map[string][]string, error) {
	return ac.GetFabricNeighborsF(nd)
}

func (ac *ApicClientMocks) GetLatestFaults(c string) ([]ApicMoAttributes, error) {
	return ac.GetLatestFaultsF(c)
}

func (ac *ApicClientMocks) SubscribeClassWebSocket(c string) (string, error) {
	return "", nil
}

func (ac *ApicClientMocks) RefreshSubscriptionWebSocket(id string) error {
	return nil
}

func (ac *ApicClientMocks) GetIp() string {
	return "1.2.3.4"
}

func (ac *ApicClientMocks) GetToken() string {
	return "aRanDoMtokEn"
}

func (ac *ApicClientMocks) Login() error {
	return nil
}

func (ac *ApicClientMocks) GetLatestEvents(c string, usr ...string) ([]ApicMoAttributes, error) {
	return ac.GetLatestEventsF(c)
}
