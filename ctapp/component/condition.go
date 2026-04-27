package component

type Condition struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	Init        Status `json:"init"`
	Run         Status `json:"run"`
	Shutdown    Status `json:"shutdown"`
	Healthcheck Status `json:"healthcheck"`
}

type StageStatus struct {
	// Group  string
	// Name   string
	Stage  Stage
	Status Status
}

// Healthcheck stage excluded
func (c Condition) GetUnfinishedStages() []StageStatus {
	res := make([]StageStatus, 0, 3)
	if c.Init != StatusUndefined && c.Init != StatusDone {
		res = append(res, StageStatus{
			// Group:  c.Group,
			// Name:   c.Name,
			Stage:  StageInit,
			Status: c.Init,
		})
	}
	if c.Run != StatusUndefined && c.Run != StatusDone {
		res = append(res, StageStatus{
			// Group:  c.Group,
			// Name:   c.Name,
			Stage:  StageRun,
			Status: c.Run,
		})
	}
	if c.Shutdown != StatusUndefined && c.Shutdown != StatusDone {
		res = append(res, StageStatus{
			// Group:  c.Group,
			// Name:   c.Name,
			Stage:  StageShutdown,
			Status: c.Shutdown,
		})
	}
	return res
}
