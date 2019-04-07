package gmx

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type IGetter interface {
	Get() (string, error)
}
type FuncGetter func() (string, error)
func (this FuncGetter) Get() (string, error) {
	return this()
}

type ISetter interface {
	Set(val string) error
}
type FuncSetter func(string) error
func (this FuncSetter) Set(val string) error {
	return this(val)
}

type ICaller interface {
	Call(val... string) (string, error)
}
type FuncCaller func(val... string) (string, error)
func (this FuncCaller)Call(val... string) (string, error) {
	return this(val...)
}

type MXItem struct {
	Name   string
	Getter IGetter
	Setter ISetter
	Caller ICaller
}

func NewMXItem(name string, getter IGetter, setter ISetter, caller ICaller) *MXItem {
	return &MXItem{Name: name, Getter: getter, Setter: setter, Caller: caller}
}

func NewMXItemIns(name string, ins interface{}) (*MXItem, error) {
	rType := reflect.TypeOf(ins)
	rValue := reflect.ValueOf(ins)

	for rType.Kind() == reflect.Ptr {
		if rValue.IsNil() {
			return nil, fmt.Errorf("nil interface with name %v", name)
		}
		rType = rType.Elem()
		rValue = rValue.Elem()
	}
	rValue = Hack.ValuePatchWrite(rValue)
	if rType.Kind() == reflect.Func {
		switch rType.NumOut() {
		case 0:
		case 1:
		case 2:
			if rType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
				return nil, fmt.Errorf("return type error")
			}
		default:
			return nil, fmt.Errorf("caller %v num out must less than 2", name)
		}
		caller := FuncCaller(func(val... string) (string, error) {
			if len(val) != rType.NumIn() {
				return "", fmt.Errorf("argument num not match %v:%v", len(val), rType.NumIn())
			}
			params := make([]reflect.Value, len(val), len(val))
			for i:=0; i<len(params); i++ {
				params[i] = reflect.New(rType.In(i)).Elem()
				err := StringPropertyInjects.Inject(params[i], val[i])
				if err != nil {
					return "", err
				}
			}
			rv := rValue.Call(params)
			switch rType.NumOut() {
			case 0:
				return "", nil
			case 1:
				return rv[0].Interface().(string), nil
			case 2:
				return rv[0].Interface().(string), rv[1].Interface().(error)
			}
			return "", errors.New("unknown")
		})
		return NewMXItem(name, nil, nil, caller), nil
	} else {
		getter := FuncGetter(func() (string, error) {
			if rType.Kind() == reflect.Struct {
				val, err := json.Marshal(ins)
				return string(val), err
			} else {
				return fmt.Sprintf("%v", ins), nil
			}
		})
		var setter ISetter
		if rValue.CanAddr() {
			setter = FuncSetter(func(val string) error {
				return StringPropertyInjects.Inject(rValue, val)
			})
		}
		return NewMXItem(name, getter, setter, nil), nil
	}
}