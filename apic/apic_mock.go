package apic

import "log"

type ApicClientMocks struct {
	tkn     string
	baseURL string
}

// TODO: check if this approach is valid
var (
	ApicMockClient ApicClientMocks
	loginF         func() error
	GetProcEntityF func() []ApicMoAttributes
	GetEnpointF    func(mac string) []ApicMoAttributes
)

// Mock functions default values
func (ac *ApicClientMocks) SetDefaultFunctions() {
	loginF = func() error {
		log.Println("Mock: APIC Login")
		return nil
	}

	GetProcEntityF = func() []ApicMoAttributes {
		procs := []ApicMoAttributes{}
		procs = append(procs, ApicMoAttributes{"dn": "abc", "cpuPct": "50", "memFree": "0"})
		procs = append(procs, ApicMoAttributes{"dn": "def", "cpuPct": "40", "memFree": "10"})
		return procs
	}

	GetEnpointF = func(mac string) []ApicMoAttributes {
		eps := []ApicMoAttributes{}
		return eps
	}
}

func (ac *ApicClientMocks) login() error {
	return loginF()
}

func (ac *ApicClientMocks) GetProcEntity() []ApicMoAttributes {
	return GetProcEntityF()
}

func (ac *ApicClientMocks) GetEnpoint(mac string) []ApicMoAttributes {
	return GetEnpointF(mac)
}
