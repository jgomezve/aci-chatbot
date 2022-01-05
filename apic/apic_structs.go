package apic

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
