package apic

type procEntity struct {
	CpuPct  string `json:"cpuPct"`
	MemFree string `json:"memFree"`
	Dn      string `json:"dn"`
}

type aaaLogin struct {
	Token string `json:"token"`
}
