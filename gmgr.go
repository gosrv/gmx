package gmx

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type IToString interface {
	ToString(ins interface{}) (string, error)
}
type FuncToString func(ins interface{}) (string, error)

func (this FuncToString) ToString(ins interface{}) (string, error) {
	return this(ins)
}

type IFromString interface {
	FromString(val reflect.Value, str string) error
}
type FuncFromString func(val reflect.Value, str string) error

func (this FuncFromString) FromString(val reflect.Value, str string) error {
	return this(val, str)
}

type MXManager struct {
	Items      map[string]*MXItem
	ToString   map[reflect.Type]IToString
	FromString map[reflect.Type]IFromString
}

func NewMXManager() *MXManager {
	mgr := &MXManager{
		Items:      make(map[string]*MXItem),
		ToString:   map[reflect.Type]IToString{},
		FromString: map[reflect.Type]IFromString{},
	}
	mgr.FromString[nil] = FuncFromString(StringPropertyInjects.Inject)
	mgr.ToString[nil] = FuncToString(func(ins interface{}) (string, error) {
		rType := reflect.TypeOf(ins)
		rValue := reflect.ValueOf(ins)
		for rValue.Kind() == reflect.Ptr {
			if rValue.IsNil() {
				return "", fmt.Errorf("nil interface")
			}
			rType = rType.Elem()
			rValue = rValue.Elem()
		}

		fmt.Println(rValue)
		fmt.Println(rType)
		fmt.Println(ins)
		if rValue.Kind() == reflect.Struct {
			val, err := json.Marshal(rValue.Interface())
			return string(val), err
		} else {
			return fmt.Sprintf("%v", rValue.Interface()), nil
		}
	})
	return mgr
}

func (this *MXManager) SetToString(tp reflect.Type, toString IToString) {
	this.ToString[tp] = toString
}

func (this *MXManager) GetToString(tp reflect.Type) IToString {
	toStr, ok := this.ToString[tp]
	if ok {
		return toStr
	}
	return this.ToString[nil]
}

func (this *MXManager) SetFromString(tp reflect.Type, fromString IFromString) {
	this.FromString[tp] = fromString
}

func (this *MXManager) GetFromString(tp reflect.Type) IFromString {
	fromStr, ok := this.FromString[tp]
	if ok {
		return fromStr
	}
	return this.FromString[nil]
}

func (this *MXManager) AddItemIns(name string, ins interface{}) error {
	if ins == nil {
		return fmt.Errorf("manage nil item %v", name)
	}

	item, err := NewMXItemIns(name, ins, this)
	if err != nil {
		return err
	}
	return this.AddItem(item)
}

func (this *MXManager) AddItemOpt(name string, getter IGetter, setter ISetter) error {
	return this.AddItem(NewMXItem(name, getter, setter, nil))
}

func (this *MXManager) AddCaller(name string, caller ICaller) error {
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
