package object

import "encoding/json"

type Param struct {
	RegistryAddr string
	Namespace     string
	RepoName    string
	Label         string
	SysCode       string
}

func (t *Param)SelectSysCode()  {
	if t.SysCode == "" {
		t.SysCode=t.Namespace
	}
}
func (t Param) MarshalBinary() (data []byte, err error) {
	return json.Marshal(t)
}
func (t *Param) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, t)
}
