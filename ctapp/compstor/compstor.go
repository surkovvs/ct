package compstor

import (
	"errors"
	"slices"
	"sync"

	"github.com/surkovvs/ct/ctapp/component"
)

var (
	ErrGroupAlreadyRegistered     = errors.New("group already registered")
	ErrGroupNotFound              = errors.New("group not found")
	ErrComponentAlreadyRegistered = errors.New("component already registered")
)

type groupNum int

type CompsStorage struct {
	mu           *sync.Mutex
	groups       map[string]SequentialGroup
	groupCounter groupNum
}

func NewCompsStorage() CompsStorage {
	return CompsStorage{
		mu:     &sync.Mutex{},
		groups: make(map[string]SequentialGroup),
	}
}

func (cs *CompsStorage) AddComponent(groupName string, comp component.Comp) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	group, ok := cs.groups[groupName]
	if !ok {
		group = SequentialGroup{
			name: groupName,
			num:  cs.groupCounter,
		}
		cs.groupCounter++
	}
	group.comps = append(group.comps, comp)
	cs.groups[groupName] = group

	return nil
}

func (cs *CompsStorage) AddGroup(groupName string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	_, ok := cs.groups[groupName]
	if ok {
		return ErrGroupAlreadyRegistered
	}

	cs.groups[groupName] = SequentialGroup{
		name: groupName,
		num:  cs.groupCounter,
	}
	cs.groupCounter++
	return nil
}

func (cs *CompsStorage) GetOrderedGroupList() []SequentialGroup {
	cs.mu.Lock()
	groupList := make([]SequentialGroup, 0, len(cs.groups))
	for _, group := range cs.groups {
		groupList = append(groupList, group)
	}
	cs.mu.Unlock()
	slices.SortFunc(groupList, func(a, b SequentialGroup) int {
		return int(a.num - b.num)
	})
	return groupList
}

func (cs *CompsStorage) GetGroupByName(name string) (SequentialGroup, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	group, ok := cs.groups[name]
	if !ok {
		return group, ErrGroupNotFound
	}
	return group, nil
}

func (cs *CompsStorage) GetUnsortedHealthcheckers() []component.Comp {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	compList := make([]component.Comp, 0)
	for _, group := range cs.groups {
		for _, comp := range group.comps {
			if comp.IsHealthchecker() {
				compList = append(compList, comp)
			}
		}
	}

	return compList
}

func (cs *CompsStorage) GetUnsortedShutdowners() []component.Comp {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	compList := make([]component.Comp, 0)
	for _, group := range cs.groups {
		for _, comp := range group.comps {
			if comp.IsShutdowner() {
				compList = append(compList, comp)
			}
		}
	}

	return compList
}

func (cs *CompsStorage) GetConditions() []component.Condition {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	var conds []component.Condition
	for _, group := range cs.groups {
		for _, comp := range group.comps {
			conds = append(conds, comp.GetCondition())
		}
	}
	return conds
}
