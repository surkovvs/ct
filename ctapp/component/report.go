package component

type ReportCode int

const (
	CodeEmpty = iota
	CodeGSDrecall
	CodeInfo
	CodeError
)

type Report struct {
	Err     error
	group   string
	module  string
	message string
	Code    ReportCode
}

func (r Report) String() string {
	var pref string
	if len(r.group) != 0 && len(r.module) != 0 {
		pref = "group '" + r.group + "', module '" + r.module + "', "
	}
	switch {
	case len(r.message) != 0 && r.Err != nil:
		return pref + r.message + ": " + r.Err.Error()
	case len(r.message) != 0:
		return pref + r.message
	case r.Err != nil:
		return pref + "error: " + r.Err.Error()
	}
	return ""
}
