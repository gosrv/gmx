package gmx

import (
	"fmt"
)

type MXManager struct {
	Items map[string]*MXItem
}

func NewMXManager() *MXManager {
	return &MXManager{
		Items:make(map[string]*MXItem),
	}
}

func (this *MXManager)AddItemIns(name string, ins interface{}) error {
	if ins == nil {
		return fmt.Errorf("manage nil item %v", name)
	}

	item, err := NewMXItemIns(name, ins)
	if err != nil {
		return err
	}
	return this.AddItem(item)
}

func (this *MXManager) AddItemOpt(name string, getter IGetter, setter ISetter) error {
	return this.AddItem(NewMXItem(name, getter, setter, nil))
}

func (this *MXManager) AddItemCaller(name string, caller ICaller) error {
	return this.AddItem(NewMXItem(name, nil, nil, caller))
}

func (this *MXManager) AddItem(item *MXItem) error {
	if item == nil {
		return fmt.Errorf("manage nil item")
	}
	if _, ok := this.Items[item.Name]; ok {
		return fmt.Errorf("duplicate manage item name %v", item.Name)
	}
	this.Items[item.Name] = item
	return nil
}
