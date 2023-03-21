package generator

import (
	"fmt"
	"strconv"
	"strings"

	e "github.com/fefit/fet/lib/expression"
	t "github.com/fefit/fet/types"
	"github.com/fefit/fet/utils"
)

// Node for type alias
type Node = e.Node

// GenConf of generator
type GenConf struct {
	Ucfirst  bool
	AutoRoot bool
}

// GenOptions for generator
type GenOptions struct {
	Exp  *e.Expression
	NsFn t.NamespaceFn
	Str  *strings.Builder
}

// ParseOptions for generator
type ParseOptions struct {
	NoObjectIndex bool
	IsInCapture   bool
	Conf          *t.FetConfig
	Captures      *map[string]string
}

// Generator for parse code
type Generator struct {
	Conf *GenConf
}

type opFnNames map[string]string

const (
	// SPACE CONSTANT
	SPACE           = " "
	toFloatFn       = "INJECT_TO_FLOAT"
	toFloatOrString = "INJECT_TO_FORS"
	indexFn         = "INJECT_INDEX"
	concatFn        = "concat"
)

var (
	compareFnNames = opFnNames{
		">=": "ge",
		">":  "gt",
		"==": "eq",
		"<":  "lt",
		"<=": "le",
		"!=": "ne",
	}
	operatorFnNames = func() opFnNames {
		ops := opFnNames{
			"||":    "or",
			"&&":    "and",
			"!":     "not",
			"+":     "INJECT_PLUS",
			"-":     "INJECT_MINUS",
			"*":     "INJECT_MULTIPLE",
			"/":     "INJECT_DIVIDE",
			"%":     "INJECT_MOD",
			"&":     "INJECT_BITAND",
			"bitor": "INJECT_BITOR",
			"^":     "INJECT_BITXOR",
			"**":    "INJECT_POWER",
		}
		for key, name := range compareFnNames {
			ops[key] = name
		}
		return ops
	}()
	// LiteralSymbols for keyword
	LiteralSymbols = map[string]string{
		"true":  "true",
		"false": "false",
		"null":  "nil",
		"nil":   "nil",
	}
	// NoNeedIndexFuncs funcs
	NoNeedIndexFuncs = map[string]bool{
		"empty": true,
		"isset": true,
	}
)

// Build for code
func (gen *Generator) Build(node *Node, options *GenOptions, parseOptions *ParseOptions) (result string, noDelimit bool, err error) {
	// conf := gen.Conf
	var str strings.Builder
	options.Str = &str
	if noDelimit, err = gen.parseRecursive(node, options, parseOptions); err != nil {
		return "", noDelimit, err
	}
	return str.String(), noDelimit, nil
}

func (gen *Generator) wrapToFloat(node *Node, options *GenOptions, parseOptions *ParseOptions, op string) error {
	str := options.Str
	isNative := false
	fn := toFloatFn
	if _, ok := compareFnNames[op]; ok {
		isNative = true
		// if equal or not equal
		if op == "==" || op == "!=" {
			fn = toFloatOrString
		}
	}
	if isNative {
		str.WriteString("(")
		str.WriteString(fn)
		str.WriteString(SPACE)
	}
	_, err := gen.parseRecursive(node, options, parseOptions)
	if isNative {
		str.WriteString(")")
	}
	return err
}

// FieldType for identifier
type FieldType int

// ObjectRoot fieldtypes
const (
	ObjectRoot FieldType = iota
	ObjectField
	FuncName
	ExpName
)

// parse identifier
func (gen *Generator) parseIdentifier(options *GenOptions, parseOptions *ParseOptions, name string, fieldType FieldType) error {
	nsFn, str := options.NsFn, options.Str
	conf := gen.Conf
	isInCapture, parseConf := parseOptions.IsInCapture, parseOptions.Conf
	if val, ok := LiteralSymbols[name]; ok {
		if fieldType != ExpName {
			panic(fmt.Sprint("syntax error: unexpect token ", name))
		} else {
			str.WriteString(val)
		}
	} else {
		origName := name
		isVar, name := nsFn(name)
		if isVar {
			if isInCapture {
				str.WriteString("$.Variables.")
				varName := strings.TrimPrefix(name, "$")
				str.WriteString(utils.Ucase(varName))
			} else {
				str.WriteString(name)
			}
		} else {
			if fieldType != FuncName {
				isRootField := fieldType == ObjectRoot || fieldType == ExpName
				if isRootField {
					if !utils.IsIdentifier(origName, parseConf.Mode) {
						return fmt.Errorf("wrong identifier name: %s", origName)
					}
					if !isInCapture && name == "ROOT" {
						str.WriteString("$")
						return nil
					}
				}
				if isInCapture {
					str.WriteString("$.")
					str.WriteString("Data.")
				} else if conf.AutoRoot {
					// trait with root data
					str.WriteString("$.")
				} else {
					// trait with variables
					str.WriteString("$")
				}
				if conf.Ucfirst {
					name = utils.Ucase(name)
				}
			}
			str.WriteString(name)
		}
	}
	return nil
}

func (gen *Generator) parseRecursive(node *Node, options *GenOptions, parseOptions *ParseOptions) (noDelimit bool, err error) {
	str, exp := options.Str, options.Exp
	noObjectIndex, parseConf, captures := parseOptions.NoObjectIndex, parseOptions.Conf, parseOptions.Captures
	curType := node.Type
	conf := gen.Conf
	isUnaryNot := node.Operator == "!"
	if isUnaryNot {
		str.WriteString("(not ")
	}
	if curType == "raw" {
		token := node.Token
		switch t := token.(type) {
		case *e.StringToken:
			stat := t.Stat
			vars := t.Variables
			runes := stat.Values
			if len(vars) > 0 {
				i, total := 1, len(runes)
				str.WriteString("(" + concatFn + SPACE)
				for _, pos := range vars {
					if pos.StartIndex > i {
						text := string(runes[i:pos.StartIndex])
						str.WriteString(" \"")
						str.WriteString(text)
						str.WriteString("\" ")
					}
					express := string(runes[pos.StartIndex+1 : pos.EndIndex-1])
					ast, _ := exp.Parse(express)
					var inner string
					if inner, noDelimit, err = gen.Build(ast, options, parseOptions); err != nil {
						return noDelimit, err
					}
					str.WriteString(inner)
					i = pos.EndIndex
					if i >= total {
						break
					}
				}
				if i < total-1 {
					str.WriteString(" \"")
					str.WriteString(string(runes[i:]))
				}
				str.WriteString(")")
			} else {
				str.WriteString(string(runes))
			}
		case *e.NumberToken:
			str.WriteString(strconv.FormatFloat(t.ToNumber(), 'f', -1, 64))
		case *e.IdentifierToken:
			stat := t.Stat
			name := string(stat.Values)
			if name == "$fet" {
				str.WriteString(".")
			} else if err = gen.parseIdentifier(options, parseOptions, name, ExpName); err != nil {
				return noDelimit, err
			}
		}
	} else if curType == "object" {
		args := node.Arguments
		total := len(args)
		root := node.Root
		isParsed := false
		isStatic := false
		addIndexFn := func() {
			if !noObjectIndex {
				str.WriteString("(" + indexFn + " ")
			}
		}
		if root.Type == "raw" {
			if t, ok := root.Token.(*e.IdentifierToken); ok {
				rootName := string(t.Stat.Values)
				if rootName == "$fet" {
					isStatic = true
					isParsed = true
					isStaticOk := true
					names := []string{}
					for i := 0; i < total; i++ {
						cur := args[i]
						isIdent := false
						if cur.Type == "raw" {
							if t, ok := cur.Token.(*e.IdentifierToken); ok {
								isIdent = true
								names = append(names, string(t.Stat.Values))
							}
						}
						if !isIdent {
							isStaticOk = false
							break
						}
					}
					if isStaticOk {
						count := len(names)
						if count == 1 {
							switch names[0] {
							case "now":
								str.WriteString("now")
							case "debug":
								noDelimit = true
								if parseConf.Debug {
									str.WriteString(`<script>(function(){try{var data = JSON.parse("{{json_encode $}}");console.log(data);window.__DEBUG__=data;}catch(e){}})();</script>`)
								}
							default:
								panic("unsupport static variable $fet." + names[0])
							}
						} else if count == 2 {
							first, second := names[0], names[1]
							if first == "config" {
								switch second {
								case "leftDelimiter":
									str.WriteString("\"" + parseConf.LeftDelimiter + "\"")
								case "rightDelimiter":
									str.WriteString("\"" + parseConf.RightDelimiter + "\"")
								case "compileDir":
									str.WriteString("\"" + parseConf.CompileDir + "\"")
								case "templateDir":
									str.WriteString("\"" + parseConf.TemplateDir + "\"")
								default:
									// isStaticOk = false
								}
							} else if first == "capture" {
								keyName := "$fet.capture." + second
								if variable, ok := (*captures)[keyName]; ok {
									str.WriteString("template \"$capture_" + second + "\" " + variable)
								} else {
									panic("unfined capture:" + keyName)
								}
							} else {
								panic("wrong static variable $fet." + first + "." + second)
							}
						} else {
							panic("unexpected static variable $fet")
						}
					}
				} else {
					addIndexFn()
					if err = gen.parseIdentifier(options, parseOptions, string(t.Stat.Values), ObjectRoot); err != nil {
						return noDelimit, err
					}
					isParsed = true
				}
			}
		}
		if !isParsed {
			addIndexFn()
			if noDelimit, err = gen.parseRecursive(root, options, parseOptions); err != nil {
				return noDelimit, err
			}
		}
		if !isStatic {
			str.WriteString(SPACE)
			for i := 0; i < total; i++ {
				cur := args[i]
				curType := cur.Type
				str.WriteString(SPACE)
				if curType == "raw" {
					token := cur.Token
					if t, ok := token.(*e.StringToken); ok {
						prop := string(t.Stat.Values)
						if conf.Ucfirst {
							str.WriteString(utils.Ucase(prop))
						} else {
							str.WriteString(prop)
						}
					} else if t, ok := token.(*e.NumberToken); ok {
						index := t.ToNumber()
						str.WriteString(strconv.FormatInt(int64(index), 10))
					} else if t, ok := token.(*e.IdentifierToken); ok {
						ident := string(t.Stat.Values)
						if cur.Operator == "." {
							if conf.Ucfirst {
								ident = utils.Ucase(ident)
							}
							str.WriteString("\"")
							str.WriteString(ident)
							str.WriteString("\"")
						} else {
							if err = gen.parseIdentifier(options, parseOptions, ident, ObjectField); err != nil {
								return noDelimit, err
							}
						}
					} else {
						if noDelimit, err = gen.parseRecursive(cur, options, parseOptions); err != nil {
							return noDelimit, err
						}
					}
				} else {
					if noDelimit, err = gen.parseRecursive(cur, options, parseOptions); err != nil {
						return noDelimit, err
					}
				}
			}
		}
		if !noObjectIndex && !isStatic {
			str.WriteString(")")
		}
		if !noObjectIndex {
			parseOptions.NoObjectIndex = false
		}
	} else if curType == "function" {
		root := node.Root
		args := node.Arguments
		str.WriteString("(")
		isParsed := false
		if root.Type == "raw" {
			if t, ok := root.Token.(*e.IdentifierToken); ok {
				name := string(t.Stat.Values)
				if err = gen.parseIdentifier(options, parseOptions, name, FuncName); err != nil {
					return noDelimit, err
				}
				if _, ok := NoNeedIndexFuncs[name]; ok {
					parseOptions.NoObjectIndex = true
				}
				isParsed = true
			}
		}
		if !isParsed {
			if noDelimit, err = gen.parseRecursive(root, options, parseOptions); err != nil {
				return noDelimit, err
			}
		}
		str.WriteString(SPACE)
		for i, total := 0, len(args); i < total; i++ {
			if i > 0 {
				str.WriteString(SPACE)
			}
			if noDelimit, err = gen.parseRecursive(args[i], options, parseOptions); err != nil {
				return noDelimit, err
			}
		}
		str.WriteString(")")
		// reset no object index
		if parseOptions.NoObjectIndex {
			parseOptions.NoObjectIndex = false
		}
	} else {
		op := node.Operator
		if name, ok := operatorFnNames[op]; ok {
			str.WriteString("(")
			str.WriteString(name)
			str.WriteString(SPACE)
		}
		if err = gen.wrapToFloat(node.Left, options, parseOptions, op); err != nil {
			return noDelimit, err
		}
		str.WriteString(SPACE)
		if err = gen.wrapToFloat(node.Right, options, parseOptions, op); err != nil {
			return noDelimit, err
		}
		str.WriteString(")")
	}
	if isUnaryNot {
		str.WriteString(")")
	}
	return noDelimit, nil
}

// New for Generator
func New(conf *GenConf) *Generator {
	return &Generator{
		Conf: conf,
	}
}
