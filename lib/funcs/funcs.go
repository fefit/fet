package funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fefit/dateutil"
)

type OperatorNumberFn func(interface{}, interface{}) interface{}
type OperatorIntFn func(int64, int64) int64
type ResultNumberFn func(args ...interface{}) interface{}
type ResultIntFn func(args ...interface{}) int64
type JSON map[string]interface{}
type LoopChan struct {
	Chan chan int
	Loop int
}
type CaptureData struct {
	Variables map[string]interface{}
	Data      interface{}
}

func (lc *LoopChan) init() {
	lc.Chan = make(chan int, 1)
	lc.Loop = -1
	_, _ = lc.Next()
}
func (lc *LoopChan) Close() (string, error) {
	lc.Loop = -1
	close(lc.Chan)
	return "", nil
}
func (lc *LoopChan) Next() (string, error) {
	lc.Loop++
	lc.Chan <- lc.Loop
	return "", nil
}

// All combine
func All() template.FuncMap {
	injects := Inject()
	helpers := Helpers()
	for key, fn := range injects {
		helpers[key] = fn
	}
	return helpers
}

// Inject funcs
func Inject() template.FuncMap {
	injects := template.FuncMap{}
	injects["INJECT_PLUS"] = generateNumberFunc(func(a, b interface{}) interface{} {
		if a, b, err := toIntNumbers(a, b); err == nil {
			return a + b
		}
		if a, b, err := toFloatNumbers(a, b); err == nil {
			return a + b
		} else {
			panic(makeHaltInfo("plus(+)", err))
		}
	}, true)
	injects["INJECT_MINUS"] = generateNumberFunc(func(a, b interface{}) interface{} {
		if a, b, err := toIntNumbers(a, b); err == nil {
			return a - b
		}
		if a, b, err := toFloatNumbers(a, b); err == nil {
			return a - b
		} else {
			panic(makeHaltInfo("minus(-)", err))
		}
	}, true)
	injects["INJECT_MULTIPLE"] = generateNumberFunc(func(a, b interface{}) interface{} {
		if a, b, err := toIntNumbers(a, b); err == nil {
			return a * b
		}
		if a, b, err := toFloatNumbers(a, b); err == nil {
			return a * b
		} else {
			panic(makeHaltInfo("multiple(*)", err))
		}
	}, true)
	injects["INJECT_DIVIDE"] = generateNumberFunc(func(a, b interface{}) interface{} {
		if a, b, err := toFloatNumbers(a, b); err == nil {
			return a / b
		} else {
			panic(makeHaltInfo("divide(/)", err))
		}
	}, false)
	injects["INJECT_MOD"] = generateNumberFunc(func(a, b interface{}) interface{} {
		if a, b, err := toFloatNumbers(a, b); err == nil {
			return math.Mod(a, b)
		} else {
			panic(makeHaltInfo("mod(%)", err))
		}
	}, false)
	injects["INJECT_POWER"] = generateNumberFunc(func(a, b interface{}) interface{} {
		if a, b, err := toFloatNumbers(a, b); err == nil {
			return math.Pow(a, b)
		} else {
			panic(makeHaltInfo("power(**)", err))
		}
	}, false)
	injects["INJECT_BITAND"] = generateIntFunc(func(a, b int64) int64 {
		return a & b
	})
	injects["INJECT_BITOR"] = generateIntFunc(func(a, b int64) int64 {
		return a | b
	})
	injects["INJECT_BITXOR"] = generateIntFunc(func(a, b int64) int64 {
		return a ^ b
	})
	injects["INJECT_TO_FLOAT"] = toFloat
	injects["INJECT_TO_FORS"] = toFloatOrString
	injects["INJECT_MAKE_LOOP_CHAN"] = func() (*LoopChan, error) {
		loopChan := &LoopChan{}
		loopChan.init()
		return loopChan, nil
	}
	injects["INJECT_INDEX"] = index
	injects["INJECT_CAPTURE_SCOPE"] = capture
	return injects
}

// Helpers funcs
func Helpers() template.FuncMap {
	helpers := template.FuncMap{}
	// output
	helpers["safe"] = safe
	// maths
	helpers["ceil"] = ceil
	helpers["floor"] = floor
	helpers["min"] = generateNumberFunc(func(a, b interface{}) interface{} {
		if a, b, err := toIntNumbers(a, b); err == nil {
			if a < b {
				return a
			}
			return b
		}
		if a, b, err := toFloatNumbers(a, b); err == nil {
			if a < b {
				return a
			}
			return b
		} else {
			panic(makeHaltInfo("min", err))
		}
	}, true)
	helpers["max"] = generateNumberFunc(func(a, b interface{}) interface{} {
		if a, b, err := toIntNumbers(a, b); err == nil {
			if a > b {
				return a
			}
			return b
		}
		if a, b, err := toFloatNumbers(a, b); err == nil {
			if a > b {
				return a
			}
			return b
		} else {
			panic(makeHaltInfo("max", err))
		}
	}, true)
	// format
	helpers["number_format"] = numberFormat
	// strings
	helpers["truncate"] = truncate
	helpers["concat"] = concat
	helpers["ucwords"] = strings.Title
	helpers["trim"] = trim
	helpers["strtolower"] = strings.ToLower
	helpers["strtoupper"] = strings.ToUpper
	// assert
	helpers["empty"] = empty
	// date
	helpers["now"] = now
	helpers["strtotime"] = dateutil.StrToTime
	helpers["date_format"] = dateutil.DateFormat
	// helper
	helpers["count"] = count
	helpers["mrange"] = makeRange
	helpers["json_encode"] = jsonEncode
	// slice, don't add this line since go1.13
	helpers["slice"] = slice
	return helpers
}

func safe(html string) template.HTML {
	return template.HTML(html)
}

var floatType = reflect.TypeOf(float64(0))
var intType = reflect.TypeOf(int64(0))

func toFloat(num interface{}) (float64, error) {
	switch t := num.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int16:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case uint:
		return float64(t), nil
	case uint16:
		return float64(t), nil
	case uint32:
		return float64(t), nil
	case uint64:
		return float64(t), nil
	default:
		v := reflect.ValueOf(num)
		v = reflect.Indirect(v)
		if !v.Type().ConvertibleTo(floatType) {
			return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
		}
		fv := v.Convert(floatType)
		return fv.Float(), nil
	}
}

func toInt(num interface{}) (int64, error) {
	switch t := num.(type) {
	case int64:
		return t, nil
	case float64:
		return int64(t), nil
	case int:
		return int64(t), nil
	case float32:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case uint:
		return int64(t), nil
	case uint16:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		return int64(t), nil
	default:
		v := reflect.ValueOf(num)
		v = reflect.Indirect(v)
		if !v.Type().ConvertibleTo(intType) {
			return 0, fmt.Errorf("cannot convert %v to int", v.Type())
		}
		fv := v.Convert(intType)
		return fv.Int(), nil
	}
}

func toFloatOrString(target interface{}) (interface{}, error) {
	switch t := target.(type) {
	case string:
		return t, nil
	default:
		return toFloat(target)
	}
}

func trim(args ...interface{}) string {
	argsNum := len(args)
	if argsNum > 0 {
		if target, ok := args[0].(string); ok {
			chars := ` \t\n\r\0\x0B`
			if argsNum == 2 {
				if trims, ok := args[1].(string); ok {
					chars = trims
				}
			}
			return strings.Trim(target, chars)
		}
	}
	return ""
}

func isInteger(target interface{}) bool {
	switch target.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case complex64, complex128:
		// ignore complex
	}
	return false
}

func toIntNumbers(a, b interface{}) (int64, int64, error) {
	var err error
	if a, ok := a.(int64); ok {
		if b, ok := b.(int64); ok {
			return a, b, nil
		}
		err = fmt.Errorf("the second argument '%v' is not an int64 type", b)
	} else {
		err = fmt.Errorf("the first argument '%v' is not an int64 type", a)
	}
	return 0, 0, err
}

func toFloatNumbers(a, b interface{}) (float64, float64, error) {
	var err error
	if a, ok := a.(float64); ok {
		if b, ok := b.(float64); ok {
			return a, b, nil
		}
		err = fmt.Errorf("the second argument '%v' is not a float64 type", b)
	} else {
		err = fmt.Errorf("the first argument '%v' is not a float64 type", a)
	}
	return 0.0, 0.0, err
}

func makeHaltInfo(name string, err error) string {
	return fmt.Sprintf("'%s' method params error:%s", name, err.Error())
}

func generateNumberFunc(fn OperatorNumberFn, allowInt bool) (res ResultNumberFn) {
	var calc ResultNumberFn
	calc = func(args ...interface{}) interface{} {
		argsNum := len(args)
		if argsNum <= 1 {
			panic("wrong arguments")
		}
		var (
			err    error
			result interface{}
		)
		f, s := args[0], args[1]
		if allowInt && isInteger(f) && isInteger(s) {
			// when both integer, do not convert to float
			var (
				first  int64
				second int64
			)
			if first, err = toInt(f); err != nil {
				panic(err)
			}
			if second, err = toInt(s); err != nil {
				panic(err)
			}
			result = fn(first, second)
		} else {
			var (
				first  float64
				second float64
			)
			if first, err = toFloat(f); err != nil {
				panic(err)
			}
			if second, err = toFloat(s); err != nil {
				panic(err)
			}
			result = fn(first, second)
		}
		if argsNum > 2 {
			args[1] = result
			return calc(args[1:]...)
		}
		return result
	}
	return calc
}

func generateIntFunc(fn OperatorIntFn) (res ResultIntFn) {
	var calc ResultIntFn
	calc = func(args ...interface{}) int64 {
		argsNum := len(args)
		if argsNum <= 1 {
			panic("wrong arguments")
		}
		var (
			err    error
			first  int64
			second int64
		)
		f, s := args[0], args[1]
		if first, err = toInt(f); err != nil {
			panic(err)
		}
		if second, err = toInt(s); err != nil {
			panic(err)
		}
		result := fn(first, second)
		if argsNum > 2 {
			args[1] = result
			return calc(args[1:]...)
		}
		return result
	}
	return calc
}

func ceil(num float64) float64 {
	return math.Ceil(num)
}

func floor(num float64) float64 {
	return math.Floor(num)
}

func numberFormat(args ...interface{}) string {
	decimals, dot, thousandsSep := 0, ".", ","
	argsNum := len(args)
	if argsNum == 0 {
		panic("wrong arguments")
	}
	var (
		err    error
		num    float64
		prefix string
		suffix string
	)
	first := args[0]
	if num, err = toFloat(first); err != nil {
		panic(err)
	}
	if argsNum > 1 {
		if dn, ok := args[1].(int); ok && dn > 0 {
			decimals = dn
		}
	}
	if argsNum > 2 {
		if ds, ok := args[2].(string); ok {
			dot = ds
		}
	}
	if argsNum > 3 {
		if ts, ok := args[3].(string); ok {
			thousandsSep = ts
		}
	}
	numstr := strconv.FormatFloat(num, 'f', -1, 64)
	isInt := false
	dotIndex := strings.Index(numstr, ".")
	if dotIndex < 0 {
		isInt = true
		dotIndex = len(numstr)
	}
	prefix = numstr[:dotIndex]
	pres := []rune(prefix)
	total := len(pres)
	splitNum := 3
	modNum := total%splitNum - 1
	if modNum < 0 {
		modNum = 2
	}
	result := []rune{}
	sep := []rune(thousandsSep)
	for i := 0; i < total-1; i++ {
		result = append(result, pres[i])
		if i%3 == modNum {
			result = append(result, sep...)
		}
	}
	result = append(result, pres[total-1])
	if decimals > 0 {
		result = append(result, []rune(dot)...)
		if !isInt {
			suffix = numstr[dotIndex+1:]
			sufs := []rune(suffix)
			if decimals <= len(sufs) {
				result = append(result, sufs[:decimals]...)
			} else {
				zeros := strings.Repeat("0", decimals-len(sufs))
				zs := []rune(zeros)
				result = append(result, sufs...)
				result = append(result, zs...)
			}
		} else {
			zeros := strings.Repeat("0", decimals)
			zs := []rune(zeros)
			result = append(result, zs...)
		}
	}
	return string(result)
}

func truncate(content string, length int) string {
	cont := []rune(content)
	total := len(cont)
	suffix := "..."
	if length >= total {
		return content
	}
	return string(cont[:length]) + suffix
}

func makeRange(s, e interface{}, args ...interface{}) []float64 {
	step := 1.0
	if start, err := toFloat(s); err != nil {
		panic(makeHaltInfo("mrange", err))
	} else {
		if end, err := toFloat(e); err != nil {
			panic(makeHaltInfo("mrange", err))
		} else {
			if len(args) == 1 {
				if curStep, ok := args[0].(float64); ok && curStep != 0.0 {
					step = curStep
				}
			}
			result := []float64{
				start,
			}
			total := math.Floor((end - start) / step)
			needLast := true
			if start+total*step == end {
				needLast = false
			}
			for i := 1.0; i <= total; i++ {
				result = append(result, start+step*i)
			}
			if needLast {
				result = append(result, end)
			}
			return result
		}
	}
}

func capture(data interface{}, variables ...interface{}) CaptureData {
	result := CaptureData{
		Data: data,
	}
	vars := map[string]interface{}{}
	count := len(variables)
	if count > 0 && count%2 == 0 {
		for i := 0; i < count; {
			key := variables[i]
			value := variables[i+1]
			if t, ok := key.(string); ok {
				vars[t] = value
			}
			i += 2
		}
	}
	result.Variables = vars
	return result
}

func now() int64 {
	t := time.Now()
	return t.Unix()
}

func stringify(target interface{}) template.HTML {
	result, err := json.Marshal(target)
	if err != nil {
		panic(err)
	}
	return template.HTML(result)
}

func jsonEncode(str string, args ...interface{}) JSON {
	if len(args) == 1 {
		fns := template.FuncMap{
			"stringify": stringify,
		}
		tmpl, err := template.New("").Funcs(fns).Parse(str)
		if err != nil {
			panic(err)
		}
		buf := &bytes.Buffer{}
		err = tmpl.Execute(buf, args[0])
		if err != nil {
			panic(err)
		}
		str = buf.String()
	}
	result := JSON{}
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		panic(err)
	}
	return result
}

func concat(str string, args ...interface{}) string {
	var builder strings.Builder
	builder.WriteString(str)
	for _, cur := range args {
		if cur, ok := cur.(string); ok {
			builder.WriteString(cur)
		}
	}
	return builder.String()
}

func chainObject(target interface{}, args ...interface{}) (finded bool, value interface{}, err error) {
	argsNum := len(args)
	if target == nil {
		if argsNum > 0 {
			return false, nil, fmt.Errorf("can not get field of nil")
		}
		return true, nil, nil
	}
	if argsNum == 0 {
		return true, target, nil
	}
	firstArg := args[0]
	nextArgs := args[1:]
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	kind := v.Kind()
	isIntKey := func(kind reflect.Kind) bool {
		switch kind {
		case reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32:
			return true
		}
		return false
	}
	getIntKey := func(key interface{}) (int64, error) {
		switch key := key.(type) {
		case int, float64, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32:
			return toInt(key)
		}
		return 0, fmt.Errorf("it's not an integer type")
	}
	if kind == reflect.Struct {
		if key, ok := firstArg.(string); ok {
			return chainObject(v.FieldByName(key).Interface(), nextArgs...)
		}
		return false, nil, fmt.Errorf("the struct field must be string type")
	} else if kind == reflect.Map {
		mapKeyType := v.Type().Key().Kind()
		mapKeys := v.MapKeys()
		if mapKeyType == reflect.String {
			if key, ok := firstArg.(string); ok {
				for _, mv := range mapKeys {
					if name, ok := mv.Interface().(string); ok && name == key {
						target = v.MapIndex(mv).Interface()
						return chainObject(target, nextArgs...)
					}
				}
			}
		} else if isIntKey(mapKeyType) {
			if index, err := getIntKey(firstArg); err == nil {
				for _, mv := range mapKeys {
					if index == mv.Int() {
						target = v.MapIndex(mv).Interface()
						return chainObject(target, nextArgs...)
					}
				}
			}
		}
		return false, nil, fmt.Errorf("the map does not has key %v", firstArg)
	} else if kind == reflect.Slice || kind == reflect.Array {
		if index, err := getIntKey(firstArg); err == nil {
			idx := int(index)
			if idx >= 0 && idx < v.Len() {
				return chainObject(v.Index(idx).Interface(), nextArgs...)
			}
		}
	}
	return false, nil, fmt.Errorf("unsupport type")
}

func index(target interface{}, args ...interface{}) interface{} {
	finded, value, err := chainObject(target, args...)
	if err != nil || !finded {
		return nil
	}
	return value
}

func empty(target interface{}, args ...interface{}) bool {
	finded, value, err := chainObject(target, args...)
	if err != nil || !finded {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == "" || v == "0"
	case int, int32, int8, int16, int64, uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float64, float32:
		return v == 0.0
	case bool:
		return !v
	case nil:
		return true
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	}
	return false
}

func count(target interface{}, args ...interface{}) int {
	if len(args) > 0 {
		panic("the 'count' function can only have one param")
	}
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	kind := v.Kind()
	if kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array {
		return v.Len()
	} else if kind == reflect.String {
		vi := v.String()
		return len([]rune(vi))
	}
	panic("the 'count' function can only used for types 'map,array,slice,string' ")
}

/**
 * slice function
 * since go1.13 has preinclude this function,you don't need add it to the func list.
 */
func toIntList(args ...interface{}) (result []int, err error) {
	value := args[0]
	if v, ok := value.(int); ok {
		result = append(result, v)
	} else {
		var v int64
		if v, err = toInt(value); err == nil {
			result = append(result, int(v))
		} else {
			return result, err
		}
	}
	if len(args) > 1 {
		var list []int
		if list, err = toIntList(args[1:]...); err == nil {
			result = append(result, list...)
		}
	}
	return result, err
}

func slice(target interface{}, args ...interface{}) interface{} {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	kind := v.Kind()
	switch kind {
	case reflect.Array, reflect.String, reflect.Slice:
	default:
		panic("the 'count' function can only used for types 'array,slice,string'")
	}
	var (
		startIndex, endIndex, lastIndex int
		isSlice3                        bool
		err                             error
		indexs                          []int
	)
	switch len(args) {
	case 0:
		startIndex, endIndex = 0, v.Len()
	case 1:
		if index, ok := args[0].(int); ok {
			startIndex, endIndex = index, v.Len()
		} else if index, err := toInt(args[0]); err == nil {
			startIndex, endIndex = int(index), v.Len()
		}
	case 2:
		if indexs, err = toIntList(args...); err == nil {
			startIndex, endIndex = indexs[0], indexs[1]
		} else {
			panic(err)
		}
	case 3:
		if kind == reflect.String {
			panic("can't use slice3 for string type")
		} else {
			isSlice3 = true
			if indexs, err = toIntList(args...); err == nil {
				startIndex, endIndex, lastIndex = indexs[0], indexs[1], indexs[2]
			} else {
				panic(err)
			}
		}
	default:
		panic("too much arguments for slice function")
	}
	if isSlice3 {
		return v.Slice3(startIndex, endIndex, lastIndex).Interface()
	}
	return v.Slice(startIndex, endIndex).Interface()
}
