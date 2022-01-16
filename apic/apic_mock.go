//+build !apic

package apic

type ApicClientMocks struct {
	GetProcEntityF        func() ([]ApicMoAttributes, error)
	GetEnpointF           func(mac string) []ApicMoAttributes
	GetFabricInformationF func() (FabricInformation, error)
}

// TODO: check if this approach is valid
var (
	ApicMockClient ApicClientMocks
)

// Mock functions default values
func (ac *ApicClientMocks) SetDefaultFunctions() {
	ac.GetProcEntityF = func() ([]ApicMoAttributes, error) {
		procs := []ApicMoAttributes{}
		procs = append(procs, ApicMoAttributes{"dn": "abc", "cpuPct": "50", "memFree": "0"})
		procs = append(procs, ApicMoAttributes{"dn": "def", "cpuPct": "40", "memFree": "10"})
		return procs, nil
	}

	ac.GetEnpointF = func(mac string) []ApicMoAttributes {
		eps := []ApicMoAttributes{}
		return eps
	}

	ac.GetFabricInformationF = func() (FabricInformation, error) {
		return FabricInformation{Name: "Test Fabric", Health: "95"}, nil
	}
}

func (ac *ApicClientMocks) GetProcEntity() ([]ApicMoAttributes, error) {
	return ac.GetProcEntityF()
}

func (ac *ApicClientMocks) GetEnpoint(mac string) []ApicMoAttributes {
	return ac.GetEnpointF(mac)
}

func (ac *ApicClientMocks) GetFabricInformation() (FabricInformation, error) {
	return ac.GetFabricInformationF()
}
