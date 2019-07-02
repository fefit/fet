package fet

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/fefit/fet/lib/expression"
	"github.com/fefit/fet/lib/funcs"
	"github.com/fefit/fet/lib/generator"
	"github.com/fefit/fet/types"
	"github.com/fefit/fet/utils"
)

// Type type
type Type int

// Runes type
type Runes []rune

// MatchTagFn func
type MatchTagFn func(strs *Runes, index int, total int) (int, bool)

// ValidateFn func
type ValidateFn func(node *Node, conf *Config) string

// Mode for parser type
type Mode = types.Mode

// Config for FetConfig
type Config = types.FetConfig

// Params struct
type Params struct {
	startTagBeginChar string
	endTagBeginChar   string
	matchStartTag     MatchTagFn
	matchEndTag       MatchTagFn
}

// Prop of Props
type Prop struct {
	Raw  string
	Type string
}

// Props of tag
type Props map[string]*Prop

// Data struct
type Data struct {
	Name  string
	Props Props
}

// Indexs for tags
type Indexs = types.Indexs

// Quote struct
type Quote struct {
	Indexs
}

// Position struct
type Position struct {
	LineNo    int
	LineIndex int
}

// CompileOptions struct
type CompileOptions struct {
	File         string
	ParentScopes []string
	ParentNS     string
	LocalScopes  *[]string
	Includes     *[]string
	Extends      *[]string
	Captures     *map[string]string
	LocalNS      string
	ParseOptions *generator.ParseOptions
}

// Node struct
type Node struct {
	Parent       *Node
	Pair         *Node
	Type         Type
	Name         string
	Content      string
	Pwd          string
	Props        *Props
	IsClosed     bool
	Features     []*Node
	Childs       []*Node
	Current      *Node
	Quotes       []*Quote
	Context      *Runes
	GlobalScopes []string
	LocalScopes  []string
	Fet          *Fet
	Data         *map[string][]string
	Indexs
	*Position
}

// NodeSets for nodelist
type NodeSets map[string][]*Node

// NodeList for
type NodeList struct {
	Queues   []*Node
	Specials NodeSets
}

var (
	defConfig = &Config{
		LeftDelimiter:  "{%",
		RightDelimiter: "%}",
		CommentSymbol:  "*",
		TemplateDir:    "templates",
		CompileDir:     "templates_c",
		CompileOnline:  false,
		LowerField:     false,
		Ignores:        []string{"inc/*"},
		Mode:           types.Smarty,
	}
	supportTags = map[string]Type{
		"include": SingleType,
		"extends": SingleType,
		"for":     BlockStartType,
		"foreach": BlockStartType,
		"if":      BlockStartType,
		"elseif":  BlockFeatureType,
		"else":    BlockFeatureType,
		"block":   BlockStartType,
		"capture": BlockStartType,
	}
	validateFns = map[string]ValidateFn{
		"if":      validIfTag,
		"else":    validElseTag,
		"elseif":  validElseifTag,
		"for":     validForTag,
		"foreach": validForTag,
		"block":   validBlockTag,
		"include": validIncludeTag,
		"extends": validExtendsTag,
		"capture": validCaptureTag,
	}
)

// UnknownType need parse
const (
	UnknownType Type = iota
	TextType
	CommentType
	OutputType
	AssignType
	SingleType
	BlockType
	BlockStartType
	BlockFeatureType
	BlockEndType
)

// AddFeature method for Node
func (node *Node) AddFeature(feature *Node) {
	feature.Parent = node
	node.Current = feature
	node.Features = append(node.Features, feature)
}

// Compile method for Node
func (node *Node) Compile(options *CompileOptions) (result string, err error) {
	fet := node.Fet
	exp, gen, conf, delimit := fet.exp, fet.gen, fet.Config, fet.wrapCode
	includes, extends, captures := options.Includes, options.Extends, options.Captures
	parseOptions := options.ParseOptions
	name, content := node.Name, node.Content
	parentScopes, localScopes := options.ParentScopes, options.LocalScopes
	parentNS, localNS := options.ParentNS, options.LocalNS
	copyScopes := append([]string{}, *localScopes...)
	currentScopes := append(copyScopes, node.GlobalScopes...)
	addVarPrefix := "$"
	isSmartyMode := conf.Mode == types.Smarty
	if isSmartyMode {
		addVarPrefix = ""
	}
	namespace := func(name string) (bool, string) {
		if contains(currentScopes, name) {
			return true, addVarPrefix + name + localNS
		} else if contains(parentScopes, name) {
			return true, addVarPrefix + name + parentNS
		}
		return false, strings.TrimPrefix(name, "$")
	}
	genOptions := &generator.GenOptions{
		NsFn: namespace,
		Exp:  exp,
	}
	toError := func(err error) error {
		return node.halt(err.Error())
	}
	switch node.Type {
	case CommentType:
		// output nothing
	case TextType:
		result = strings.TrimSpace(content)
		if conf.LeftDelimiter != "{{" {
			rule := regexp.MustCompile(`(\{{2,})`)
			result = rule.ReplaceAllString(result, `{{"$1"}}`)
		}
	case AssignType, OutputType:
		isAssign := node.Type == AssignType
		if isAssign && !utils.IsIdentifier(name, conf.Mode) {
			err = node.halt("syntax error: wrong variable name '%s', please check the parser mode", name)
			break
		}
		ast, expErr := exp.Parse(content)
		if expErr != nil {
			err = toError(expErr)
		} else {
			if isAssign {
				if _, ok := generator.LiteralSymbols[name]; ok {
					err = node.halt("syntax error: can not set literal '%s' as a variable name", name)
					break
				}
				symbol := " := "
				if contains(currentScopes, name) {
					symbol = " = "
				}
				result = delimit(addVarPrefix + name + localNS + symbol + gen.Build(ast, genOptions, parseOptions))
			} else {
				result = delimit(gen.Build(ast, genOptions, parseOptions))
			}
		}
	case SingleType:
		if name == "include" {
			tpl := fet.getRealTplPath(node.Content, path.Join(node.Pwd, ".."))
			ctx := md5.New()
			ctx.Write([]byte(tpl))
			curNS := hex.EncodeToString(ctx.Sum(nil))
			if contains(*includes, tpl) || contains(*extends, tpl) {
				err = node.halt("the include file '%s' has a loop dependence", tpl)
			} else {
				incCaptures := &map[string]string{}
				incOptions := &CompileOptions{
					ParentNS:     localNS,
					LocalNS:      "_" + curNS,
					ParentScopes: currentScopes[:],
					LocalScopes:  &[]string{},
					Extends:      extends,
					Includes:     includes,
					Captures:     incCaptures,
					ParseOptions: &generator.ParseOptions{
						Conf:     conf,
						Captures: incCaptures,
					},
				}
				if incResult, incErr := fet.compileFileContent(tpl, incOptions); incErr != nil {
					err = toError(incErr)
				} else {
					result = incResult
				}
			}
		} else if name == "extends" {
			// ignore extends
			// load file because of variable scopes
		}
	case BlockStartType:
		if name == "for" || name == "foreach" {
			props := *node.Props
			if props["type"].Raw == "foreach" {
				target := props["list"].Raw
				ast, expErr := exp.Parse(target)
				if expErr != nil {
					err = toError(expErr)
				} else {
					code := gen.Build(ast, genOptions, parseOptions)
					key := props["key"].Raw
					result = "range "
					if key != "" {
						result += addVarPrefix + key + localNS + ", "
					}
					result += addVarPrefix + props["value"].Raw + localNS + " := " + code
					result = delimit(result)
				}
			} else {
				data := *node.Data
				vars := data["Vars"]
				initial := data["Initial"]
				res := strings.Builder{}
				// add if block for variable context
				res.WriteString(delimit("if true"))
				for key, name := range vars {
					ast, expErr := exp.Parse(initial[key])
					if expErr != nil {
						err = toError(expErr)
						break
					}
					res.WriteString(delimit(addVarPrefix + name + localNS + ":=" + gen.Build(ast, genOptions, parseOptions)))
				}
				suffixNS := indexString(node.StartIndex) + "_" + indexString(node.EndIndex) + localNS
				chanName := "$loop_" + suffixNS
				res.WriteString(delimit(chanName + " := (INJECT_MAKE_LOOP_CHAN)"))
				res.WriteString(delimit("range " + chanName + ".Chan"))
				// Add condition code
				conds := data["Conds"][0]
				// add initial declares
				currentScopes = append(currentScopes, vars...)
				ast, expErr := exp.Parse(conds)
				if expErr != nil {
					err = toError(expErr)
				} else {
					res.WriteString(delimit("if " + gen.Build(ast, genOptions, parseOptions)))
					res.WriteString(delimit(chanName + ".Next"))
					res.WriteString(delimit("else"))
					res.WriteString(delimit(chanName + ".Close"))
					res.WriteString(delimit("end"))
					res.WriteString(delimit("if (gt " + chanName + ".Loop -1)"))
				}
				result = res.String()
			}
		} else if name == "if" {
			ast, expErr := exp.Parse(content)
			if expErr != nil {
				err = toError(expErr)
			} else {
				code := gen.Build(ast, genOptions, parseOptions)
				result = delimit("if " + code)
			}
		} else if name == "capture" {
			parseOptions.IsInCapture = true
			keyName := "$fet.capture." + content
			if _, ok := (*captures)[keyName]; ok {
				// capture name has exists
				err = node.halt("repeated capture name '" + content + "'")
			} else {
				capVar := "$fet_capture_" + content + localNS
				result = "{{" + capVar + " := (INJECT_CAPTURE_SCOPE . "
				for _, varName := range currentScopes {
					varName = strings.TrimPrefix(varName, "$")
					result += "\"" + strings.Title(varName) + "\" $" + varName
				}
				result += ")}}"
				(*captures)[keyName] = capVar
				result += "{{ define \"" + content + "\"}}"
			}
		}
	case BlockFeatureType:
		if name == "elseif" {
			ast, expErr := exp.Parse(content)
			if expErr != nil {
				err = toError(expErr)
			} else {
				code := gen.Build(ast, genOptions, parseOptions)
				result = delimit("else if " + code)
			}
		} else if name == "else" {
			result = delimit("else")
		}
	case BlockEndType:
		pair := node.Pair
		if name == "block" {
			blockScopes := pair.LocalScopes
			if len(blockScopes) > 0 {
				*options.LocalScopes = append(*options.LocalScopes, blockScopes...)
			}
		} else if name == "for" {
			props := *pair.Props
			if props["type"].Raw == "for" {
				data := *pair.Data
				// close index condition
				result += delimit("end")
				loops := data["Loops"]
				for i, total := 0, len(loops); i < total; {
					ast, expErr := exp.Parse(loops[i])
					if expErr != nil {
						err = toError(expErr)
						break
					}
					code := gen.Build(ast, genOptions, parseOptions)
					ast, expErr = exp.Parse(loops[i+1])
					if expErr != nil {
						err = toError(expErr)
						break
					}
					code += " = " + gen.Build(ast, genOptions, parseOptions)
					result += delimit(code)
					i += 2
				}
				if err == nil {
					//  first: close range; last: close if
					result += delimit("end") + delimit("end")
				}
			} else {
				result = delimit("end")
			}
		} else {
			if name == "capture" {
				parseOptions.IsInCapture = false
			}
			result = delimit("end")
		}
	default:
		// types not assign
	}
	return
}

// halt errors
func (node *Node) halt(format string, args ...interface{}) error {
	var errmsg string
	if node.Pwd != "" {
		errmsg = "[file:'" + node.Pwd + "']"
	}
	errmsg += "[line:" + indexString(node.LineNo) + ",col:" + indexString(node.StartIndex-node.LineIndex+1) + "]" + fmt.Sprintf(format, args...)
	return errors.New(errmsg)
}

func getPrevFeature(node *Node) (feature *Node) {
	nodes := node.Features
	total := len(nodes)
	if total > 2 {
		feature = nodes[total-2]
	}
	return
}

/**
 * validators
 */
func validIfPrevCondOk(node *Node) (errmsg string) {
	prev := getPrevFeature(node.Parent)
	if prev == nil {
		return
	} else if prev.Name != "elseif" {
		errmsg = "the prev tag of \"" + node.Name + "\" should not be \"" + prev.Name + "\""
	}
	return
}
func validIfRoot(node *Node) (errmsg string) {
	if node.Parent != nil {
		errmsg = "the \"" + node.Name + "\" tag should not appears in " + node.Parent.Name
	}
	return
}
func validIfBlockCorrect(node *Node, blockName string) (errmsg string) {
	pname := node.Parent.Name
	if pname != blockName {
		errmsg = "the \"" + node.Name + "\" tag can not used in block \"" + pname + "\""
	}
	return
}
func validIfOnlyOneStrParam(node *Node) (errmsg string) {
	if len(node.Quotes) != 1 {
		errmsg = "wrong tag \"" + node.Name + "\""
	} else {
		quote := node.Quotes[0]
		all := string((*node.Context)[quote.StartIndex:quote.EndIndex])
		if all != node.Content {
			errmsg = "wrong file string \"" + all + "\" of " + node.Name + " tag"
		} else {
			content := []rune(node.Content)
			node.Content = string(content[1 : len(content)-1])
		}
	}
	return
}

// tag validators
func validIfTag(node *Node, conf *Config) (errmsg string) {
	if node.Content == "" {
		errmsg = "the \"if\" tag does not have a condition expression"
	}
	return
}
func validElseTag(node *Node, conf *Config) (errmsg string) {
	if node.Content != "" {
		errmsg = "the \"else\" tag should not use a condition expression"
	} else if errmsg = validIfPrevCondOk(node); errmsg == "" {
		errmsg = validIfBlockCorrect(node, "if")
	}
	return
}
func validElseifTag(node *Node, conf *Config) (errmsg string) {
	if node.Content == "" {
		errmsg = "the \"elseif\" tag does not have a condition expression"
	} else if errmsg = validIfPrevCondOk(node); errmsg == "" {
		errmsg = validIfBlockCorrect(node, "if")
	}
	return
}
func validBlockTag(node *Node, conf *Config) (errmsg string) {
	block := node.Parent
	if block != nil && block.Parent != nil {
		errmsg = "the \"block\" tag should be root tag,can not appears in \"" + block.Parent.Name + "\""
	} else {
		errmsg = validIfOnlyOneStrParam(node)
	}
	return
}

func validCaptureTag(node *Node, conf *Config) (errmsg string) {
	capture := node.Parent
	if capture != nil && capture.Parent != nil {
		errmsg = "the \"capture\" tag should be root tag,can not appears in \"" + capture.Parent.Name + "\""
	} else {
		errmsg = validIfOnlyOneStrParam(node)
	}
	return
}

func validForTag(node *Node, conf *Config) (errmsg string) {
	name := node.Name
	content := node.Content
	runes := Runes(node.Content)
	segs := []string{}
	total := 0
	maxNum := 2
	prevIsSpace := true
	hasKey := false
	normalErr := "the \"" + content + "\" tag is not correct"
	isSmartyMode := conf.Mode == types.Smarty
	var (
		list, key, value *Prop
	)
	isForEach := false
	if name == "foreach" {
		if isSmartyMode {
			isForEach = true
			count := len(runes)
			isArrow, prevIsArrow := false, false
			for count > 0 {
				count--
				s := runes[count]
				if unicode.IsSpace(s) {
					prevIsSpace = true
					if total >= maxNum {
						segs = append(segs, string(runes[:count]))
						break
					}
				} else {
					if s == '>' && count >= 1 && runes[count-1] == '=' {
						hasKey = true
						maxNum += 2
						isArrow = true
						prevIsArrow = true
						count--
					}
					if prevIsSpace || isArrow || prevIsArrow {
						if isArrow {
							segs = append(segs, "=>")
							isArrow = false
						} else {
							segs = append(segs, string(s))
							if prevIsArrow {
								prevIsArrow = false
							}
						}
						total++
					} else {
						segs[total-1] = string(s) + segs[total-1]
					}
					prevIsSpace = false
				}
			}
			total = len(segs)
			if total != maxNum+1 {
				errmsg = normalErr
			} else {
				if (hasKey && segs[1] != "=>") || segs[total-2] != "as" {
					return "wrong syntax \"foreach\" block, please check the compile mode"
				}
				list = &Prop{
					Raw: segs[total-1],
				}
				if !hasKey {
					key = &Prop{
						Raw: "",
					}
				} else {
					key = &Prop{
						Raw: segs[total-3],
					}
				}
				value = &Prop{
					Raw: segs[0],
				}
			}
		} else {
			return "the Gofet mode does not support 'foreach' block, use 'for' instead"
		}
	} else {
		needTryForIn := true
		if isSmartyMode || len(strings.Split(content, ";")) >= 3 {
			needTryForIn = false
			i := 0
			num := len(runes)
			colon := ';'
			part := []string{}
			allParts := [][][]string{}
			isInCont := false
			isInTranslate := false
			isInQuote := false
			bLevel := 0
			addParts := func(part []string) {
				if len(allParts) <= total {
					allParts = append(allParts, [][]string{})
				}
				allParts[total] = append(allParts[total], part)
			}
			for i < num && total < 3 {
				s := runes[i]
				ch := string(s)
				count := len(part)
				i++
				if isInQuote || isInTranslate {
					// check if is quote or translate
					part[count-1] += ch
					if isInTranslate {
						isInTranslate = false
					} else {
						if s == '"' {
							isInQuote = false
						} else if s == '\\' {
							isInTranslate = true
						}
					}
				} else {
					if unicode.IsSpace(s) {
						// empty
						if isInCont {
							part[count-1] += ch
						}
						continue
					}
					if s == '"' {
						isInQuote = true
					} else if s == '(' {
						bLevel++
					} else if s == ')' {
						bLevel--
					} else if bLevel == 0 {
						isColon := s == colon
						if isColon || s == ',' {
							addParts(part)
							if isColon {
								total++
							}
							part = []string{}
							isInCont = false
							continue
						} else {
							if total == 0 {
								if s == '=' {
									part = append(part, "=")
									isInCont = false
									continue
								}
							} else if total == 2 {
								if s == '-' || s == '+' {
									if i >= num {
										errmsg = "wrong iterator done"
										break
									} else {
										next := runes[i]
										if next == '=' || next == s {
											part = append(part, ch+string(next))
											i++
										} else {
											errmsg = "wrong iterator done"
											break
										}
									}
									isInCont = false
									continue
								}
							}
						}
					}
					if !isInCont {
						part = append(part, ch)
						isInCont = true
					} else {
						part[count-1] += ch
					}
				}

			}
			if len(part) > 0 {
				addParts(part)
			}
			if len(allParts) == 3 {
				if errmsg != "" {
					return errmsg
				}
				node.Data = &map[string][]string{}
				data := *node.Data
				initial, conds, lastRuns := allParts[0], allParts[1], allParts[2]
				// parse initials
				vars := []string{}
				declares := []string{}
				for _, part := range initial {
					name := strings.TrimSpace(part[0])
					if len(part) != 3 || part[1] != "=" || !utils.IsIdentifier(name, conf.Mode) {
						return "wrong for initial"
					}
					vars = append(vars, name)
					declares = append(declares, part[2])
				}
				data["Vars"] = vars
				data["Initial"] = declares
				// parse break or continue conds
				cond := ""
				for _, part := range conds {
					if len(part) == 1 {
						cur := strings.TrimSpace(part[0])
						if cur != "" {
							if cond != "" {
								cond += " && "
							}
							cond += cur
						}
					}
				}
				if cond == "" {
					return "wrong for condition sentence"
				}
				lastCond := []string{}
				lastCond = append(lastCond, cond)
				data["Conds"] = lastCond
				// parse loops
				loops := []string{}
				for _, part := range lastRuns {
					count := len(part)
					if count <= 1 {
						return "wrong 'for' last loop"
					}
					name, symbol := strings.TrimSpace(part[0]), part[1]
					if !utils.IsIdentifier(name, conf.Mode) {
						return "wrong for identifier '" + name + "'"
					}
					op := string(([]rune(symbol))[0])
					isStepOn := symbol == "++" || symbol == "--"
					if isStepOn || (symbol == "+=" || symbol == "-=") {
						if (isStepOn && count > 2 && strings.TrimSpace(part[2]) != "") || (!isStepOn && (count != 3 || strings.TrimSpace(part[2]) == "")) {
							return "wrong for loop"
						}
						loops = append(loops, name)
						if isStepOn {
							loops = append(loops, name+op+"1")
						} else {
							loops = append(loops, name+op+part[2])
						}
					}
				}
				data["Loops"] = loops
				props := *node.Props
				props["type"] = &Prop{
					Raw: "for",
				}
			} else {
				if isSmartyMode {
					return "wrong synatx 'for' statement, please check the parser mode"
				}
				needTryForIn = true
			}
		}
		if needTryForIn {
			isForEach = true
			isComma := false
			prevIsComma := false
			for index, s := range runes {
				if unicode.IsSpace(s) {
					prevIsSpace = true
					if total >= maxNum {
						segs = append(segs, string(runes[index+1:]))
						break
					}
				} else {
					if s == ',' {
						hasKey = true
						isComma = true
						prevIsComma = true
						maxNum += 2
					}
					if prevIsSpace || isComma || prevIsComma {
						segs = append(segs, string(s))
						total++
						if isComma {
							isComma = false
						} else if prevIsComma {
							prevIsComma = false
						}
					} else {
						segs[total-1] += string(s)
					}
					prevIsSpace = false
				}
			}
			if len(segs) != maxNum+1 {
				return normalErr
			}
			inIndex := 1
			if hasKey && segs[1] == "," {
				inIndex = 3
			}
			if segs[inIndex] != "in" {
				return "wrong syntax \"for\" block"
			}
			value = &Prop{
				Raw: segs[0],
			}
			if !hasKey {
				key = &Prop{
					Raw: "",
				}
			} else {
				key = &Prop{
					Raw: segs[inIndex-1],
				}
			}
			list = &Prop{
				Raw: segs[inIndex+1],
			}
		}
	}
	if isForEach {
		if !utils.IsIdentifier(value.Raw, conf.Mode) {
			return fmt.Sprintf("the 'for' label's value '%s' is a wrong identifier", value.Raw)
		}
		if hasKey && !utils.IsIdentifier(key.Raw, conf.Mode) {
			return fmt.Sprintf("the 'for' label's key '%s' is a wrong identifier", key.Raw)
		}
		props := *node.Props
		props["key"] = key
		props["value"] = value
		props["list"] = list
		props["type"] = &Prop{
			Raw: "foreach",
		}
	}
	return
}
func validIncludeTag(node *Node, conf *Config) (errmsg string) {
	errmsg = validIfOnlyOneStrParam(node)
	return
}
func validExtendsTag(node *Node, conf *Config) (errmsg string) {
	if node.Parent != nil {
		errmsg = "the \"extends\" tag should be root tag,can not appears in \"" + node.Parent.Name + "\""
	} else {
		errmsg = validIfOnlyOneStrParam(node)
	}
	return
}

// Validate method for Node
func (node *Node) Validate(conf *Config) (errmsg string) {
	if node.Type == UnknownType {
		errmsg = "unknown type"
	} else if !node.IsClosed {
		errmsg = "the tag is not closed"
	} else {
		name := node.Name
		if fn, exists := validateFns[name]; exists {
			errmsg = fn(node, conf)
		}
	}
	return
}

// Fet struct
type Fet struct {
	*Config
	Params
	compileDir  string
	templateDir string
	exp         *expression.Expression
	gen         *generator.Generator
	datas       map[string]interface{}
	cwd         string
	tmpl        *template.Template
}

func mergeConfig(options *Config) *Config {
	conf := &Config{}
	*conf = *defConfig
	if options.Mode > 0 {
		conf.Mode = options.Mode
	}
	if options.LeftDelimiter != "" {
		conf.LeftDelimiter = options.LeftDelimiter
	}
	if options.RightDelimiter != "" {
		conf.RightDelimiter = options.RightDelimiter
	}
	if options.TemplateDir != "" {
		conf.TemplateDir = options.TemplateDir
	}
	if options.CompileDir != "" {
		conf.CompileDir = options.CompileDir
	}
	if options.CompileOnline {
		conf.CompileOnline = true
	}
	if options.LowerField {
		conf.LowerField = true
	}
	if options.Ignores != nil {
		conf.Ignores = options.Ignores
	}
	return conf
}
func buildMatchTagFn(len int, tag *Runes) MatchTagFn {
	return func(strs *Runes, index int, total int) (int, bool) {
		if index+len > total {
			return index, false
		}
		i := 1
		for i < len {
			if (*strs)[index+i] != (*tag)[i] {
				return index, false
			}
			i++
		}
		return index + len - 1, true
	}
}

// New create a template enginner
func New(config *Config) (fet *Fet, err error) {
	config = mergeConfig(config)
	ld, rd := Runes(config.LeftDelimiter), Runes(config.RightDelimiter)
	params := Params{
		startTagBeginChar: string(ld[0]),
		endTagBeginChar:   string(rd[0]),
		matchStartTag:     buildMatchTagFn(len(ld), &ld),
		matchEndTag:       buildMatchTagFn(len(rd), &rd),
	}
	gen := generator.New(&generator.GenConf{
		Ucfirst: !config.LowerField,
	})
	exp := expression.New()
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	fet = &Fet{
		Config: config,
		Params: params,
		gen:    gen,
		exp:    exp,
		datas:  make(map[string]interface{}),
		cwd:    cwd,
	}
	tmpl := template.New("")
	tmpl = tmpl.Funcs(funcs.All())
	fet.tmpl = tmpl
	fet.compileDir = fet.getLastDir(config.CompileDir)
	fet.templateDir = fet.getLastDir(config.TemplateDir)
	fmt.Println(fet.compileDir)
	return fet, nil
}
func ltrimIndex(strs *Runes, i int, total int) int {
	index := i
	for index < total {
		if unicode.IsSpace((*strs)[index]) {
			index++
		} else {
			break
		}
	}
	return index
}

/**
 * int to string
 */
func indexString(index int) string {
	return fmt.Sprintf("%d", index)
}

/**
 * remove right spaces
 */
func rtrim(strs *Runes, start int, end int) string {
	for start < end {
		if unicode.IsSpace((*strs)[end-1]) {
			end--
		} else {
			break
		}
	}
	return string((*strs)[start:end])
}

/**
 * close node type
 */
func closeTag(node *Node, endIndex int) {
	node.IsClosed = true
	node.EndIndex = endIndex + 1
}

// add variable prefix
func (fet *Fet) addVarPrefix(name string) string {
	conf := fet.Config
	if conf.Mode == types.Smarty {
		return name
	}
	return "$" + name
}

// wrap left delimiter, right delimiter
func (fet *Fet) wrapCode(code string) string {
	return "{{" + code + "}}"
}

// parse
func (fet *Fet) parse(codes string, pwd string) (result *NodeList, err error) {
	var (
		isInComment, isTagStart, isInBlockTag, isSubTemplate bool
		node                                                 *Node
		quote                                                *Quote
		markIndex, lineNo, lineIndex, blockStartIndex        int
		blocks                                               []*Node
		queues                                               []*Node
	)
	specials := NodeSets{}
	globals := []string{}
	lineNo = 1
	getLastBlock := func() *Node {
		len := len(blocks)
		if len > 0 {
			return blocks[len-1]
		}
		return nil
	}
	setFeatureChild := func(node *Node) {
		if block := getLastBlock(); block != nil {
			current := block.Current
			node.Parent = current
		}
	}
	initToStart := func() {
		isTagStart = false
		node = nil
		markIndex = 0
	}
	isNeedScope := func() bool {
		return !isSubTemplate || isInBlockTag
	}
	addSpecial := func(name string, node *Node) {
		if _, ok := specials[name]; ok {
			specials[name] = append(specials[name], node)
		} else {
			specials[name] = []*Node{
				node,
			}
		}
	}
	popGlobals := func(block *Node) {
		prevFeature := block.Current
		locals := prevFeature.LocalScopes
		if locals != nil {
			globals = append([]string{}, globals[:len(globals)-len(locals)]...)
		}
	}
	strs := Runes(codes)
	total := len(strs)
LOOP:
	for i := 0; i < total; i++ {
		rn := strs[i]
		code := string(rn)
		// set lines and cols
		if rn == '\n' {
			lineNo++
			lineIndex = i
		}
		position := &Position{
			LineNo:    lineNo,
			LineIndex: lineIndex,
		}
		// match comment end
		if isInComment {
			if code == fet.CommentSymbol {
				if curIndex, isTagEnd := fet.matchEndTag(&strs, i+1, total); isTagEnd {
					i = curIndex
					node.EndIndex = curIndex + 1
					isInComment = false
					setFeatureChild(node)
					initToStart()
				}
			}
			continue
		}
		// match end tag
		if isTagStart {
			// if in quote
			if quote != nil {
				if code == "\\" {
					// translate
					i++
				} else if code == "\"" {
					// quote end
					quote.EndIndex = i + 1
					quote = nil
				}
			} else {
				if node.Type == UnknownType {
					if code == " " {
						node.Name = rtrim(&strs, markIndex+1, i)
						if curType, exists := supportTags[node.Name]; exists {
							node.Type = curType
							markIndex = i
							block := getLastBlock()
							name := node.Name
							switch curType {
							case BlockStartType:
								pnode := &Node{
									Indexs: Indexs{
										StartIndex: node.StartIndex,
									},
									Type:     BlockType,
									Name:     name,
									Position: position,
									Context:  &strs,
									Fet:      fet,
									Pwd:      pwd,
								}
								if name == "block" {
									addSpecial(name, node)
									isInBlockTag = true
									blockStartIndex = len(queues) - 1
								}
								pnode.AddFeature(node)
								if block != nil {
									pnode.Parent = block.Current
								}
								blocks = append(blocks, pnode)
								node.Parent = pnode
							case BlockFeatureType:
								if block != nil {
									popGlobals(block)
									node.GlobalScopes = globals
									block.AddFeature(node)
								} else {
									err = node.halt("wrong block feature without start block")
									break LOOP
								}
							default:
								isExtendsTag := name == "extends"
								if name == "include" || isExtendsTag {
									addSpecial(name, node)
									if isExtendsTag {
										isSubTemplate = true
										if len(blocks) > 0 {
											err = node.halt("wrong 'extends' tag, can not use 'extends' in block tags")
											break LOOP
										} else {
											for i, total := 0, len(queues); i < total-1; i++ {
												curNode := queues[i]
												if curNode.Type != CommentType && curNode.Type != TextType {
													err = node.halt("the 'extends' tag must appear at the top of template")
													break
												}
											}
											if err != nil {
												break LOOP
											}
										}
									}
								}
							}
						}
						continue
						// else try assign type
					} else if code == "=" {
						if node.Name == "" {
							// e.g <%a=c%>
							node.Name = string(strs[markIndex+1 : i])
						}
						ident := strings.TrimSpace(node.Name)
						if ident == "" {
							err = node.halt("the variable name is empty")
							break
						}
						if utils.IsIdentifier(ident, types.AnyMode) {
							node.Type = AssignType
							markIndex = i
							setFeatureChild(node)
							if isNeedScope() {
								// add scope variable
								name := node.Name
								block := getLastBlock()
								if block != nil {
									current := block.Current
									current.LocalScopes = append(current.LocalScopes, name)
								}
								globals = append(globals, name)
							}
							continue
						}
					}
					// else try BlockFeature type
				}
				if code == "\"" {
					quote = &Quote{}
					quote.StartIndex = i
					quote.EndIndex = i
					node.Quotes = append(node.Quotes, quote)
				} else if code == "'" {
					err = node.halt("do not use quote \"'\",use \" instead")
					break
				} else if code == fet.endTagBeginChar {
					curIndex, isTagEnd := fet.matchEndTag(&strs, i, total)
					if isTagEnd {
						// start
						node.IsClosed = true
						node.EndIndex = curIndex + 1
						isUnknownType := node.Type == UnknownType
						if node.Type == BlockEndType || isUnknownType {
							name := rtrim(&strs, markIndex+1, i)
							block := getLastBlock()
							setOutputType := func() {
								node.Type = OutputType
								node.Content = name
								setFeatureChild(node)
								initToStart()
							}
							if block == nil {
								if isUnknownType {
									if name == "" {
										err = node.halt("empty tag")
										break
									} else {
										setOutputType()
									}
								} else {
									err = node.halt("wrong end tag \"%s\"", name)
									break
								}
							} else {
								if isUnknownType {
									if curType, exists := supportTags[name]; exists && curType == BlockFeatureType {
										node.Name = name
										node.Type = BlockFeatureType
										popGlobals(block)
										block.AddFeature(node)
									} else {
										setOutputType()
									}
								} else {
									if block.Name != name {
										err = block.halt("the block tag \"%s\" is not closed", block.Name)
										break
									} else {
										node.Name = name
										closeTag(block, curIndex)
										current := block.Current
										if name == "block" {
											isInBlockTag = false
											current.Childs = queues[blockStartIndex:len(queues)]
										}
										node.Pair = current
										popGlobals(block)
										blocks = blocks[:len(blocks)-1]
									}
								}
							}
						} else {
							// validate tag
							noSpaceIndex := ltrimIndex(&strs, markIndex+1, i)
							node.Content = rtrim(&strs, noSpaceIndex, i)
							if errmsg := node.Validate(fet.Config); errmsg != "" {
								err = node.halt(errmsg)
								break
							}
							if node.Type == BlockStartType && (node.Name == "for" || node.Name == "foreach") && isNeedScope() {
								props := *node.Props
								forType := props["type"].Raw
								if forType == "foreach" {
									valueProp := props["value"]
									keyProp := props["key"]
									globals = append(globals, valueProp.Raw)
									node.LocalScopes = append(node.LocalScopes, valueProp.Raw)
									if keyProp.Raw != "" {
										globals = append(globals, keyProp.Raw)
										node.LocalScopes = append(node.LocalScopes, keyProp.Raw)
									}
								} else {
									data := *node.Data
									vars := data["Vars"]
									globals = append(globals, vars...)
									node.LocalScopes = append(node.LocalScopes, vars...)
								}
							}
							setFeatureChild(node)
						}
						i = curIndex
						// initial status
						initToStart()
					}
				}
			}
			continue
		}
		// match start tag
		if code == fet.startTagBeginChar {
			curIndex, startFlag := fet.matchStartTag(&strs, i, total)
			if startFlag {
				// set isTagStart true
				isTagStart = startFlag
				// if prev node is text node
				if node != nil && node.Type == TextType {
					node.IsClosed = true
					node.EndIndex = i
					node.Content = string(strs[node.StartIndex:i])
				}
				// judge next
				nextIndex := curIndex + 1
				if nextIndex == total {
					break
				}
				next := string(strs[nextIndex])
				node = &Node{
					Indexs: Indexs{
						StartIndex: i,
					},
					Position:     position,
					Props:        &Props{},
					Context:      &strs,
					Fet:          fet,
					GlobalScopes: globals,
					Pwd:          pwd,
				}
				if next == fet.CommentSymbol {
					// comment
					node.Type = CommentType
					isInComment = true
				} else {
					// remove unnecessary spaces
					noSpaceIndex := ltrimIndex(&strs, nextIndex, total)
					if noSpaceIndex != nextIndex {
						curIndex = noSpaceIndex - 1
						if noSpaceIndex == total {
							break
						}
						next = string(strs[noSpaceIndex])
					}
					if next == "/" {
						// end block
						node.Type = BlockEndType
						curIndex = nextIndex
					} else {
						// need parse again
						node.Type = UnknownType
					}
				}
				markIndex = curIndex
				i = curIndex
				queues = append(queues, node)
				continue
			}
			// else be text node
		}
		// not start,not end,but new start
		if node == nil {
			node = &Node{
				Type: TextType,
				Indexs: Indexs{
					StartIndex: i,
					EndIndex:   i,
				},
				Position:     position,
				Context:      &strs,
				Fet:          fet,
				GlobalScopes: globals,
				Pwd:          pwd,
			}
			setFeatureChild(node)
			queues = append(queues, node)
		}
	}
	// judge if has error
	if err != nil {
		return nil, err
	}
	// judge if block end,last node is closed.
	if block := getLastBlock(); block != nil {
		return nil, block.halt("unclosed block tag\"%s\"", block.Name)
	}
	if node != nil {
		if node.Type != TextType {
			return nil, node.halt("unclosed tag\"%s\"", node.Content)
		}
		node.IsClosed = true
		node.EndIndex = total
		node.Content = string(strs[node.StartIndex:node.EndIndex])
	}
	return &NodeList{
		Queues:   queues,
		Specials: specials,
	}, nil
}

// Display method
func (fet *Fet) Display(tpl string, data interface{}, output io.Writer) (err error) {
	conf := fet.Config
	if conf.CompileOnline {
		var result string
		if result, err = fet.Fetch(tpl, data); err == nil {
			_, err = os.Stdout.Write([]byte(result))
		}
		return err
	}
	compileFile := fet.getRealTplPath(tpl, fet.compileDir)
	if _, err = os.Stat(compileFile); err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("the compile file '%s' is not exist", compileFile)
		}
		return err
	}
	tmpl, _ := fet.tmpl.Clone()
	if buf, rErr := ioutil.ReadFile(compileFile); rErr != nil {
		err = rErr
	} else {
		t, pErr := tmpl.Parse(string(buf))
		if pErr != nil {
			err = pErr
		} else {
			err = t.Execute(output, data)
		}
	}
	return err
}

// Fetch method
func (fet *Fet) Fetch(tpl string, data interface{}) (result string, err error) {
	tmpl, _ := fet.tmpl.Clone()
	if code, cErr := fet.Compile(tpl, false); cErr != nil {
		err = cErr
	} else {
		t, pErr := tmpl.Parse(code)
		if pErr != nil {
			err = pErr
		} else {
			buf := new(bytes.Buffer)
			err = t.Execute(buf, data)
			if err == nil {
				result = buf.String()
			}
		}
	}
	return
}

func contains(arr []string, key string) bool {
	for _, cur := range arr {
		if cur == key {
			return true
		}
	}
	return false
}

func (fet *Fet) getRealTplPath(tpl string, currentDir string) string {
	if path.IsAbs(tpl) {
		return tpl
	}
	return path.Join(currentDir, tpl)
}

func (fet *Fet) parseFile(tpl string, blocks []*Node, extends *[]string, nested int) (*NodeList, bool, error) {
	if contains(*extends, tpl) {
		return nil, true, fmt.Errorf("the extends file is looped extend:%s", tpl)
	}
	isSubTemplate := nested > 0
	for {
		if _, err := os.Stat(tpl); err == nil {
			buf, err := ioutil.ReadFile(tpl)
			if err != nil {
				return nil, isSubTemplate, fmt.Errorf("read the file failure:%s", tpl)
			}
			content := string(buf)
			nl, err := fet.parse(content, tpl)
			if err != nil {
				return nil, isSubTemplate, err
			}
			specials := nl.Specials
			var curBlocks []*Node
			if bk, ok := specials["block"]; ok {
				curBlocks = bk
			}
			if exts, exists := specials["extends"]; exists {
				if curBlocks != nil {
					blocks = append(blocks, curBlocks...)
				}
				if nested == 0 {
					*extends = append(*extends, tpl)
				}
				tpl = fet.getRealTplPath(exts[0].Content, path.Join(tpl, ".."))
				nl, _, err := fet.parseFile(tpl, blocks, extends, nested+1)
				*extends = append(*extends, tpl)
				return nl, true, err
			}
			if !isSubTemplate {
				return nl, isSubTemplate, nil
			}
			namedBlocks := map[string]*Node{}
			for _, block := range blocks {
				name := block.Content
				if _, exists := namedBlocks[name]; !exists {
					namedBlocks[name] = block
				}
			}
			overides := map[string][]*Node{}
			counts := map[string]int{}
			for _, block := range curBlocks {
				name := block.Content
				if override, exists := namedBlocks[name]; exists {
					overides[name] = override.Childs
					counts[name] = len(block.Childs)
				}
			}
			queues := []*Node{}
			replaces := []*Node{}
			for index, total := 0, len(nl.Queues); index < total; index++ {
				node := nl.Queues[index]
				if node.Type == BlockStartType && node.Name == "block" {
					blockName := node.Content
					if count, ok := counts[blockName]; ok {
						index += count - 1
						replaces = overides[blockName]
						queues = append(queues, replaces...)
						continue
					}
				}
				queues = append(queues, node)
			}
			nl.Queues = queues
			return nl, isSubTemplate, nil
		} else if os.IsNotExist(err) {
			return nil, isSubTemplate, fmt.Errorf("the file '%s' is not exists", tpl)
		} else {
			return nil, isSubTemplate, err
		}
	}
}

// Compile for string
func (fet *Fet) compileFileContent(tpl string, options *CompileOptions) (string, error) {
	blocks := []*Node{}
	extends := options.Extends
	nl, _, err := fet.parseFile(tpl, blocks, extends, 0)
	if err != nil {
		return "", err
	}
	result := strings.Builder{}
	var (
		code string
		errs error
	)
	options.File = tpl
	for _, node := range nl.Queues {
		if code, errs = node.Compile(options); errs != nil {
			return "", errs
		}
		result.WriteString(code)
	}
	lastCode := result.String()
	return lastCode, nil
}

func (fet *Fet) getLastDir(dir string) string {
	if dir == "" {
		return fet.cwd
	}
	runes := []rune(dir)
	first := runes[0]
	if path.IsAbs(dir) {
		return dir
	} else if first == '.' {
		return path.Join(fet.cwd, dir)
	} else {
		lastDir, _ := filepath.Abs(dir)
		return lastDir
	}
}

// Compile file
func (fet *Fet) Compile(tpl string, writeFile bool) (string, error) {
	var (
		result string
		err    error
	)
	parentScopes := []string{}
	localScopes := []string{}
	includes := []string{}
	extends := []string{}
	captures := map[string]string{}
	tplFile := fet.getRealTplPath(tpl, fet.templateDir)
	compileFile := fet.getRealTplPath(tpl, fet.compileDir)
	parseOptions := &generator.ParseOptions{
		Conf:     fet.Config,
		Captures: &captures,
	}
	options := &CompileOptions{
		ParentScopes: parentScopes,
		LocalScopes:  &localScopes,
		Includes:     &includes,
		Extends:      &extends,
		Captures:     &captures,
		ParseOptions: parseOptions,
	}
	if writeFile {
		defer func() {
			if err != nil {
				fmt.Println("compile fail:", err.Error())
			} else {
				fmt.Println("compile success")
			}
		}()
		fmt.Println("compile file:", tplFile, "--->", compileFile)
	}
	if result, err = fet.compileFileContent(tplFile, options); err != nil {
		return "", err
	}
	if writeFile {
		dir := path.Dir(compileFile)
		if _, err = os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(dir, os.ModePerm)
			}
			if err != nil {
				return "", fmt.Errorf("can not open the compile dir:" + dir)
			}
		}
		err = ioutil.WriteFile(compileFile, []byte(result), 0644)
		if err != nil {
			return "", fmt.Errorf("compile file '" + compileFile + "' failure:" + err.Error())
		}
	}
	fmt.Println(extends, includes)
	return result, nil
}

func (fet *Fet) isIgnoreFile(tpl string) bool {
	conf := fet.Config
	ignores := conf.Ignores
	if len(ignores) == 0 {
		return false
	}
	for _, glob := range ignores {
		if ok, _ := filepath.Match(glob, tpl); ok {
			return true
		}
	}
	return false
}

// CompileAll files
func (fet *Fet) CompileAll() error {
	dir := fet.templateDir
	files := []string{}
	err := filepath.Walk(dir, func(pwd string, info os.FileInfo, err error) error {
		tpl, _ := filepath.Rel(dir, pwd)
		if err != nil {
			fmt.Println("read compile file failure:", err)
			return nil
		}
		if !info.IsDir() && !fet.isIgnoreFile(tpl) {
			files = append(files, tpl)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("sorry,fail to open the compile directory:%s", err.Error())
	}
	total := len(files)
	if total == 0 {
		return nil
	}
	if total == 1 {
		_, err := fet.Compile(files[0], true)
		return err
	}
	var (
		wg   sync.WaitGroup
		errs []string
	)
	wg.Add(total)
	go func() {
		for _, tpl := range files {
			_, err := fet.Compile(tpl, true)
			if err != nil {
				errs = append(errs, err.Error())
			}
			wg.Done()
		}
	}()
	wg.Wait()
	if errs != nil {
		return fmt.Errorf("compile file error:%s", strings.Join(errs, "\n"))
	}
	return nil
}
