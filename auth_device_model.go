package hermes

type Device struct {
	Id          int     `json:"id" hermes:"dbspace:devices"`
	Model       string  `json:"model"`
	Platform    string  `json:"platform"`
	Uuid        string  `json:"uuid"`
	Version     string  `json:"version" hermes:"editable"`
	Ip          string  `json:"ip"`
	CM_Id       string  `json:"cm_id" hermes:"editable"`
	App_Version float32 `json:"-" hermes:"index,searchable"`
}

var DeviceColl *Collection
