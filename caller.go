package gosocketio

import (
	"errors"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

type caller struct {
	Func        reflect.Value
	Args        []reflect.Type
	ArgsPresent bool
	Out         bool
}

var (
	ErrorCallerNotFunc     = errors.New("f is not function")
	ErrorCallerNot2Args    = errors.New("f should have 1 or 2 args")
	ErrorCallerMaxOneValue = errors.New("f should return not more than one value")
)

/**
Parses function passed by using reflection, and stores its representation
for further call on message or ack
*/
func newCaller(f interface{}) (*caller, error) {
	fVal := reflect.ValueOf(f)
	if fVal.Kind() != reflect.Func {
		return nil, ErrorCallerNotFunc
	}

	fType := fVal.Type()
	if fType.NumOut() > 1 {
		return nil, ErrorCallerMaxOneValue
	}

	curCaller := &caller{
		Func: fVal,
		Out:  fType.NumOut() == 1,
	}
	if fType.NumIn() == 1 {
		curCaller.Args = nil
		curCaller.ArgsPresent = false
	} else if fType.NumIn() >= 2 {
		types := make([]reflect.Type, fType.NumIn()-1)
		for i := 0; i < len(types); i++ {
			types[i] = fType.In(i + 1)
		}
		curCaller.Args = types
		curCaller.ArgsPresent = true
	} else {
		return nil, ErrorCallerNot2Args
	}

	return curCaller, nil
}

/**
returns function parameter as it is present in it using reflection
*/
func (c *caller) getArgs() []interface{} {
	vals := make([]interface{}, len(c.Args))

	for i := 0; i < len(vals); i++ {
		vals[i] = reflect.New(c.Args[i]).Interface()
	}

	return vals
}

/**
calls function with given arguments from its representation using reflection
*/
func (c *caller) callFunc(h *Channel, args ...interface{}) []reflect.Value {
	//nil is untyped, so use the default empty value of correct type
	if args == nil {
		args = c.getArgs()
	}

	a := []reflect.Value{reflect.ValueOf(h)}
	if c.ArgsPresent {
		for i, arg := range args {
			var iface interface{}

			if f, ok := arg.(float64); ok {
				switch c.Args[i].Kind() {
				case reflect.Int:
					iface = int(f)
				case reflect.Int8:
					iface = int8(f)
				case reflect.Int16:
					iface = int16(f)
				case reflect.Int32:
					iface = int32(f)
				case reflect.Int64:
					iface = int64(f)
				case reflect.Uint:
					iface = uint(f)
				case reflect.Uint8:
					iface = uint8(f)
				case reflect.Uint16:
					iface = uint16(f)
				case reflect.Uint32:
					iface = uint32(f)
				case reflect.Uint64:
					iface = uint64(f)
				case reflect.Float32:
					iface = float32(f)
				case reflect.Float64:
					iface = f
				}
			} else {
				switch c.Args[i].Kind() {
				case reflect.Struct:
					iface = reflect.New(c.Args[i]).Elem().Interface()

					if err := mapstructure.Decode(arg, &iface); err != nil {
						panic(err)
					}
				case reflect.Ptr:
					// TODO: This may not be right...
					iface = reflect.New(c.Args[i].Elem()).Interface()

					if err := mapstructure.Decode(arg, &iface); err != nil {
						panic(err)
					}
				default:
					iface = arg
				}
			}

			a = append(a, reflect.ValueOf(iface))
		}
	}

	return c.Func.Call(a)
}
