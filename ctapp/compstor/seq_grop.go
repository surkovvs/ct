package compstor

import "github.com/surkovvs/ct/ctapp/component"

type SequentialGroup struct {
	name  string
	num   groupNum
	comps []component.Comp
}

func (sg SequentialGroup) GetName() string {
	return sg.name
}

func (sg SequentialGroup) GetComponents() []component.Comp {
	return sg.comps
}
