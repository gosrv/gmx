package gmx

import (
	"encoding/json"
	"errors"
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

type IMXManager interface {
	SetToString(tp reflect.Type, toString IToString)
	GetToString(tp reflect.Type) IToString
	SetFromString(tp reflect.Type, fromString IFromString)
	GetFromString(tp reflect.Type) IFromString
	AddItemIns(name string, ins interface{}) error
	AddItemOpt(name string, getter IGetter, setter ISetter, setterType string) error
	AddCaller(name string, caller ICaller, info *CallerInfo) error
	AddItem(item *MXItem) error
}

type MXManager struct {
	Items      map[string]*MXItem
	ToString   map[reflect.Type]IToString
	FromString map[reflect.Type]IFromString
}

var _ IMXManager = (*MXManager)(nil)

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

func (this *MXManager) AddItemOpt(name string, getter IGetter, setter ISetter, setterType string) error {
	return this.AddItem(NewMXItem(name, getter, setter, setterType, nil, nil))
}

func (this *MXManager) AddCaller(name string, caller ICaller, info *CallerInfo) error {
	return this.AddItem(NewMXItem(name, nil, nil, "", caller, info))
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

func (this *MXManager) HandleKeys() ([]byte, error) {
	infos := make([]MXItemInfo, 0, len(this.Items))
	for _, v := range this.Items {
		infos = append(infos, v.Info)
	}
	return json.Marshal(infos)
}

func (this *MXManager) HandleGet(keys []string) ([]byte, error) {
	if len(keys) == 1 {
		item := this.Items[keys[0]]
		if item == nil || item.Getter == nil {
			return nil, errors.New("not exist")
		}
		val, err := item.Getter.Get()
		return []byte(val), err
	} else {
		rep := make([]string, 0, len(keys))
		for _, key := range keys {
			item := this.Items[key]
			if item == nil || item.Getter == nil {
				rep = append(rep, "")
			} else {
				val, _ := item.Getter.Get()
				rep = append(rep, val)
			}
		}
		return json.Marshal(rep)
	}
}

func (this *MXManager) HandleSet(keys []string, vals []string) ([]byte, error) {
	if len(keys) != len(vals) {
		return nil, errors.New("keys and values len not match")
	}
	num := 0
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val := vals[i]
		item := this.Items[key]
		if item.Setter == nil {
			continue
		}
		err := item.Setter.Set(val)
		if err != nil {
			continue
		}
		num++
	}
	return []byte(fmt.Sprintf("%v", num)), nil
}

func (this *MXManager) HandleCall(key string, params []string) ([]byte, error) {
	item := this.Items[key]
	if item == nil || item.Caller == nil {
		return nil, errors.New("not exist")
	}
	rep, err := item.Caller.Call(params...)
	return []byte(rep), err
}
