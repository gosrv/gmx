package gmx

import (
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
	Call(val ...string) (string, error)
}
type FuncCaller func(val ...string) (string, error)

func (this FuncCaller) Call(val ...string) (string, error) {
	return this(val...)
}

type MXItem struct {
	Name   string
	Getter IGetter
	Setter ISetter
	Caller ICaller
	Info   MXItemInfo
}

type CallerInfo struct {
	ParamTypes  []string
	ReturnTypes []string
}

func NewCallerInfo(paramTypes []string, returnTypes []string) *CallerInfo {
	return &CallerInfo{ParamTypes: paramTypes, ReturnTypes: returnTypes}
}

type MXItemInfo struct {
	Name       string
	Getter     bool
	Setter     bool
	CallerInfo *CallerInfo
}

func NewMXItemInfo(item *MXItem, callerInfo *CallerInfo) MXItemInfo {
	return MXItemInfo{Name: item.Name, Getter: item.Getter != nil,
		Setter: item.Setter != nil, CallerInfo: callerInfo}
}

func NewMXItem(name string, getter IGetter, setter ISetter, caller ICaller, callerInfo *CallerInfo) *MXItem {
	item := &MXItem{Name: name, Getter: getter, Setter: setter, Caller: caller}
	item.Info = NewMXItemInfo(item, callerInfo)
	return item
}

func NewMXItemIns(name string, ins interface{}, mgr *MXManager) (*MXItem, error) {
	rType := reflect.TypeOf(ins)
	rValue := reflect.ValueOf(ins)

	for rValue.Kind() == reflect.Ptr {
		if rValue.IsNil() {
			return nil, fmt.Errorf("nil interface with name %v", name)
		}
		rType = rType.Elem()
		rValue = rValue.Elem()
	}
	rValue = Hack.ValuePatchWrite(rValue)
	if rValue.Kind() == reflect.Func {
		switch rType.NumOut() {
		case 0:
		case 1:
		case 2:
			if rType.Out(0) == reflect.TypeOf((*error)(nil)).Elem() ||
				rType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
				return nil, fmt.Errorf("return type error, first must not error and second must error")
			}
		default:
			return nil, fmt.Errorf("caller %v num out must less than 2", name)
		}
		caller := FuncCaller(func(val ...string) (string, error) {
			if len(val) != rType.NumIn() {
				return "", fmt.Errorf("argument num not match %v:%v", len(val), rType.NumIn())
			}
			params := make([]reflect.Value, len(val), len(val))
			for i := 0; i < len(params); i++ {
				params[i] = reflect.New(rType.In(i)).Elem()
				err := mgr.GetFromString(rType.In(i)).FromString(params[i], val[i])
				if err != nil {
					return "", err
				}
			}
			rv := rValue.Call(params)
			switch rType.NumOut() {
			case 0:
				return "", nil
			case 1:
				if err, ok := rv[0].Interface().(error); ok {
					return "", err
				} else {
					return mgr.GetToString(rType.Out(0)).ToString(rv[0].Interface())
				}
			case 2:
				err := rv[1].Interface().(error)
				str, errs := mgr.GetToString(rType.Out(0)).ToString(rv[0].Interface())
				if err == nil {
					err = errs
				}
				return str, err
			}
			return "", errors.New("unknown")
		})
		pt := make([]string, 0, rType.NumIn())
		for i := 0; i < rType.NumIn(); i++ {
			pt = append(pt, rType.In(i).Name())
		}
		rt := make([]string, 0, rType.NumOut())
		for i := 0; i < rType.NumOut(); i++ {
			rt = append(rt, rType.Out(i).Name())
		}
		return NewMXItem(name, nil, nil, caller, NewCallerInfo(pt, rt)), nil
	} else {
		toString := mgr.GetToString(rType)
		getter := FuncGetter(func() (string, error) { return toString.ToString(rValue.Interface()) })
		var setter ISetter
		fromString := mgr.GetFromString(rType)
		if rValue.CanAddr() {
			setter = FuncSetter(func(val string) error {
				return fromString.FromString(rValue, val)
			})
		}
		return NewMXItem(name, getter, setter, nil, nil), nil
	}
}
