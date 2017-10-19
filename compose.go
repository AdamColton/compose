package compose

import (
	"fmt"
	"reflect"
)

type caller func([]reflect.Value) []reflect.Value

// New takes funcs as it's arugments and returns their composition. Will panic
// if there is an error.
func Must(fns ...interface{}) interface{} {
	fn, err := New(fns...)
	if err != nil {
		panic(err)
	}
	return fn
}

// New takes funcs as it's arugments and returns their composition.
func New(fns ...interface{}) (interface{}, error) {
	if len(fns) < 2 {
		return nil, fmt.Errorf("Requires at least 2 functions")
	}

	fnTypes, fnVals, err := funcTypesVals(fns...)
	if err != nil {
		return nil, err
	}

	in, out, callers, vard, err := getInOutCallers(fnTypes, fnVals)
	if err != nil {
		return nil, err
	}

	fn := func(vals []reflect.Value) []reflect.Value {
		for _, call := range callers {
			vals = call(vals)
		}
		return vals
	}
	fnType := reflect.FuncOf(in, out, vard)
	return reflect.MakeFunc(fnType, fn).Interface(), nil
}

func funcTypesVals(fns ...interface{}) ([]reflect.Type, []reflect.Value, error) {
	ts := make([]reflect.Type, len(fns))
	vs := make([]reflect.Value, len(fns))
	for i, fn := range fns {
		t := reflect.TypeOf(fn)
		if t.Kind() != reflect.Func {
			return nil, nil, fmt.Errorf("Argument %d is not a func", i)
		}
		ts[i] = t
		vs[i] = reflect.ValueOf(fn)
	}
	return ts, vs, nil
}

func getInOutCallers(fnTypes []reflect.Type, fnVals []reflect.Value) (in, out []reflect.Type, callers []caller, vard bool, err error) {
	callers = make([]caller, 1, 2*len(fnTypes))

	in, out = getInOut(fnTypes[0])
	vard = fnTypes[0].IsVariadic()
	callers[0] = getCaller(vard, fnVals[0])

	for i, fn := range fnTypes[1:] {
		var correction caller
		curIn, nextOut := getInOut(fn)

		if fn.IsVariadic() {
			correction, err = checkVariadic(curIn, out, i+1)
			curIn = curIn[:len(curIn)-1]
		} else if len(curIn) != len(out) {
			err = fmt.Errorf("Incorrect argument length at %d: return length %d != argument length %d", i+1, len(out), len(curIn))
		}
		if err != nil {
			return
		}

		for j, arg := range curIn {
			if arg != out[j] {
				err = fmt.Errorf("Argument mismatch in fn %d at %d: return type %s != argument type %s", i+1, j, out[j], arg)
				return
			}
		}

		if c := getCaller(fn.IsVariadic(), fnVals[i+1]); correction == nil {
			callers = append(callers, c)
		} else {
			callers = append(callers, correction, c)
		}

		out = nextOut
	}
	return
}

func getInOut(fn reflect.Type) ([]reflect.Type, []reflect.Type) {
	in := make([]reflect.Type, fn.NumIn())
	out := make([]reflect.Type, fn.NumOut())
	for i := range in {
		in[i] = fn.In(i)
	}
	for i := range out {
		out[i] = fn.Out(i)
	}
	return in, out
}

func getCaller(vard bool, v reflect.Value) caller {
	if vard {
		return v.CallSlice
	}
	return v.Call
}

// checkVariadic will check that the variadic argument lines up with the return
// values of the preceeding argument. There are 3 cases where the output of one
// function can be fed into the input of a variadic function that require a
// correction. Caller will hold this correction if it is required.
func checkVariadic(in, out []reflect.Type, fnIdx int) (caller, error) {
	ln := len(in) - 1
	if lo := len(out); ln == lo-1 && in[ln] == out[ln] {
		// Exact match, all good, no correction required
		return nil, nil
	} else if ln > lo { // not off-by-one; in can be 1 shorter than out
		return nil, fmt.Errorf("Too few returns preceeding %d, does not match arguments", fnIdx)
	}
	// len(out)>=len(in)-1
	// need to cast tail of out to slice
	s := in[ln]
	v := s.Elem()
	for i, t := range out[ln:] {
		if t != v {
			return nil, fmt.Errorf("Return value preceeding function %d at %d does not match variadic type", fnIdx, i)
		}
	}
	toConvert := len(out) - ln
	correction := func(vs []reflect.Value) []reflect.Value {
		sl := reflect.MakeSlice(s, toConvert, toConvert)
		for i, v := range vs[ln:] {
			sl.Index(i).Set(v)
		}
		return append(vs[:ln], sl)
	}
	return correction, nil
}
