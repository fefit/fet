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
)

type OperatorFloatFn func(float64, float64) float64
type OperatorIntFn func(int64, int64) int64
type ResultFloatFn func(args ...interface{}) float64
type ResultIntFn func(args ...interface{}) int64
type JSON map[string]interface{}
type LoopChan struct {
	Chan chan int
	Loop int
}

func (lc *LoopChan) init() {
	lc.Chan = make(chan int, 1)
	lc.Loop = -1
	lc.Next()
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
	injects["INJECT_PLUS"] = generateFloatFunc(func(a, b float64) float64 {
		return a + b
	})
	injects["INJECT_MINUS"] = generateFloatFunc(func(a, b float64) float64 {
		return a - b
	})
	injects["INJECT_MULTIPLE"] = generateFloatFunc(func(a, b float64) float64 {
		return a * b
	})
	injects["INJECT_DIVIDE"] = generateFloatFunc(func(a, b float64) float64 {
		return a / b
	})
	injects["INJECT_MOD"] = generateFloatFunc(func(a, b float64) float64 {
		return math.Mod(a, b)
	})
	injects["INJECT_POWER"] = generateFloatFunc(func(a, b float64) float64 {
		return math.Pow(a, b)
	})
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
	injects["INJECT_MAKE_LOOP_CHAN"] = func() (*LoopChan, error) {
		loopChan := &LoopChan{}
		loopChan.init()
		return loopChan, nil
	}
	return injects
}

// Helpers funcs
func Helpers() template.FuncMap {
	helpers := template.FuncMap{}
	helpers["safe"] = safe
	helpers["ceil"] = ceil
	helpers["floor"] = floor
	helpers["number_format"] = number_format
	helpers["truncate"] = truncate
	helpers["mrange"] = makeRange
	helpers["concat"] = concat
	helpers["json_encode"] = jsonEncode
	helpers["concat"] = concat
	helpers["min"] = generateFloatFunc(func(a, b float64) float64 {
		if a > b {
			return b
		}
		return a
	})
	helpers["max"] = generateFloatFunc(func(a, b float64) float64 {
		if a > b {
			return a
		}
		return b
	})
	helpers["empty"] = empty
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
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case uint:
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
	case float32:
		return int64(t), nil
	case int:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case uint:
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

func generateFloatFunc(fn OperatorFloatFn) (res ResultFloatFn) {
	var calc ResultFloatFn
	calc = func(args ...interface{}) float64 {
		argsNum := len(args)
		if argsNum <= 1 {
			panic("wrong arguments")
		}
		var (
			err    error
			first  float64
			second float64
		)
		f, s := args[0], args[1]
		if first, err = toFloat(f); err != nil {
			panic(err)
		}
		if second, err = toFloat(s); err != nil {
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

func number_format(args ...interface{}) string {
	decimals, dot, thousands_sep := 0, ".", ","
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
			thousands_sep = ts
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
	sep := []rune(thousands_sep)
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

func makeRange(start, end float64, args ...interface{}) []float64 {
	step := 1.0
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

func empty(target interface{}, args ...interface{}) bool {
	if target == nil {
		return false
	}
	switch t := target.(type) {
	case int, float64, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32:
		return t == 0
	case string:
		return t == ""
	case bool:
		return t == false
	}
	v := reflect.ValueOf(target)
	kind := v.Kind()
	argsNum := len(args)
	if kind == reflect.Map {
		if argsNum > 0 {
			firstArg := args[0]
			fmt.Println(reflect.ValueOf(firstArg).Type() == v.Type())
			// if obj, ok := target.(map[string]interface{}); ok {
			// 	if key, ok := firstArg.(string); ok {
			// 		if last, ok := obj[key]; ok {
			// 			return empty(last, args[1:]...)
			// 		}
			// 	}
			// 	return false
			// }
			// if obj, ok := target.(map[int]interface{}); ok {
			// 	if key, ok := firstArg.(int); ok {
			// 		if last, ok := obj[key]; ok {
			// 			return empty(last, args[1:]...)
			// 		}
			// 	}
			// 	return false
			// }
			// return false
		}
		return true
	}
	return false
}
