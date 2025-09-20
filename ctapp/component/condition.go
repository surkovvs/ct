package component

type Condition struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	Init        string `json:"init"`
	Run         string `json:"run"`
	Shutdown    string `json:"shutdown"`
	Healthcheck string `json:"healthcheck"`
}
