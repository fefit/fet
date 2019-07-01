package funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fefit/dateutil"
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
type CaptureData struct {
	Variables map[string]interface{}
	Data      interface{}
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
	injects["INJECT_INDEX"] = index
	injects["INJECT_CAPTURE_SCOPE"] = capture
	return injects
}

// Helpers funcs
func Helpers() template.FuncMap {
	helpers := template.FuncMap{}
	helpers["safe"] = safe
	helpers["ceil"] = ceil
	helpers["floor"] = floor
	helpers["number_format"] = numberFormat
	helpers["truncate"] = truncate
	helpers["mrange"] = makeRange
	helpers["concat"] = concat
	helpers["json_encode"] = jsonEncode
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
	helpers["count"] = count
	helpers["now"] = now
	helpers["strToTime"] = dateutil.StrToTime
	helpers["date_format"] = dateutil.DateFormat
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
func strToTime(target interface{}) (time.Time, error) {
	return time.Now(), nil
}

func dateFormat(target interface{}, format string) (string, error) {
	layouts := []string{
		// year
		"Y", "2006",
		"y", "06",
		// month
		"m", "01",
		"n", "1",
		// date
		"d", "02",
		"j", "2",
		// hours
		"h", "03",
		"g", "3",
		"G", "15",
		// minutes
		"i", "04",
		// seconds
		"s", "05",
	}
	formats := map[string]string{
		// am, pm
		"a": "pm",
		"A": "PM",
		// month
		"F": "January",
		"M": "Jan",
		// week
		"D": "Mon",
		"l": "Monday",
	}
	N := func(t time.Time) string {
		weekday := t.Weekday()
		return fmt.Sprintf("%d", int(weekday))
	}
	w := func(t time.Time) string {
		weekday := t.Weekday()
		dayNum := int(weekday) % 7
		return fmt.Sprintf("%d", dayNum)
	}
	z := func(t time.Time) string {
		yearday := t.YearDay()
		return fmt.Sprintf("%d", yearday-1)
	}
	W := func(t time.Time) string {
		_, week := t.ISOWeek()
		return fmt.Sprintf("%d", week)
	}
	L := func(t time.Time) string {
		yearday := t.YearDay()
		if yearday > 365 {
			return "1"
		}
		return "0"
	}
	t := func(t time.Time) string {
		nums := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
		monthIndex := int(t.Month()) - 1
		if monthIndex == 1 && t.YearDay() > 365 {
			return "29"
		}
		return fmt.Sprintf("%d", nums[monthIndex])
	}
	H := func(t time.Time) string {
		hour := t.Hour()
		return fmt.Sprintf("%02d", hour)
	}
	fns := map[string]func(t time.Time) string{
		"N": N,
		"w": w,
		"z": z,
		"W": W,
		"L": L,
		"t": t,
		"H": H,
	}
	repRule := strings.NewReplacer(layouts...)
	layout := repRule.Replace(format)
	regRule := func() *regexp.Regexp {
		var str strings.Builder
		str.WriteString("[")
		for key := range formats {
			str.WriteString(key)
		}
		for key := range fns {
			str.WriteString(key)
		}
		str.WriteString("]")
		rule, _ := regexp.Compile(str.String())
		return rule
	}()
	var timeTarget time.Time
	if cur, ok := target.(time.Time); ok {
		timeTarget = cur
	} else {
		if cur, err := strToTime(target); err == nil {
			timeTarget = cur
		} else {
			return "", err
		}
	}
	result := timeTarget.Format(layout)
	result = regRule.ReplaceAllStringFunc(result, func(name string) string {
		if layout, ok := formats[name]; ok {
			return timeTarget.Format(layout)
		} else if fn, ok := fns[name]; ok {
			return fn(timeTarget)
		}
		return ""
	})
	return result, nil
}

/*
* https://gist.github.com/elliotchance/d419395aa776d632d897
 */
func replaceWith(re *regexp.Regexp, str string, repl func(args ...string) string) string {
	result := ""
	lastIndex := 0
	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}
		result += str[lastIndex:v[0]] + repl(groups...)
		lastIndex = v[1]
	}
	return result + str[lastIndex:]
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
	case int, float64, uint, int64, int32, int16, int8, uint64, uint32, uint16, uint8, float32:
		return v == 0
	case bool:
		return v == false
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
	kind := v.Kind()
	if kind == reflect.Ptr {
		v = v.Elem()
		kind = v.Kind()
	}
	if kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array {
		return v.Len()
	} else if kind == reflect.String {
		vi := v.String()
		return len([]rune(vi))
	}
	panic("the 'count' function can only used in types 'map,array,slice,string' ")
}
