//go:build !apic
// +build !apic

package apic

type ApicClientMocks struct {
	GetProcEntityF          func() ([]ApicMoAttributes, error)
	GetEnpointF             func(mac string) []ApicMoAttributes
	GetFabricInformationF   func() (FabricInformation, error)
	GetEndpointInformationF func(m string) ([]EndpointInformation, error)
	GetFabricNeighborsF     func(nd string) (map[string][]string, error)
	GetLatestFaultsF        func(c string) ([]ApicMoAttributes, error)
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

	ac.GetEnpointF = func(mac string) []ApicMoAttributes {
		eps := []ApicMoAttributes{}
		return eps
	}

	ac.GetFabricInformationF = func() (FabricInformation, error) {
		return FabricInformation{Name: "Test Fabric", Health: "95"}, nil
	}

	ac.GetEndpointInformationF = func(m string) ([]EndpointInformation, error) {
		return []EndpointInformation{}, nil
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

func (ac *ApicClientMocks) WsClassSubscription(c string) (string, error) {
	return "", nil
}

func (ac *ApicClientMocks) WsSubcriptionRefresh(id string) error {
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
