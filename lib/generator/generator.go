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
	Ucfirst bool
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
	Conf          *t.FetConfig
	IsInCapture   bool
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
func (gen *Generator) Build(node *Node, options *GenOptions, parseOptions *ParseOptions) (string, error) {
	// conf := gen.Conf
	var str strings.Builder
	options.Str = &str
	if err := gen.parseRecursive(node, options, parseOptions); err != nil {
		return "", err
	}
	return str.String(), nil
}

func (gen *Generator) wrapToFloat(node *Node, options *GenOptions, parseOptions *ParseOptions, op string) error {
	str := options.Str
	isNative := false
	fn := toFloatFn
	if _, ok := compareFnNames[op]; ok {
		isNative = true
		if op == "==" {
			fn = toFloatOrString
		}
	}
	if isNative {
		str.WriteString("(")
		str.WriteString(fn)
		str.WriteString(SPACE)
	}
	err := gen.parseRecursive(node, options, parseOptions)
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
				str.WriteString(strings.Title(varName))
			} else {
				str.WriteString(name)
			}
		} else {
			if fieldType != FuncName {
				if (fieldType == ObjectRoot || fieldType == ExpName) && !utils.IsIdentifier(origName, parseConf.Mode) {
					return fmt.Errorf("wrong identifier name: %s", origName)
				}
				str.WriteString("$.")
				if isInCapture {
					str.WriteString("Data.")
				}
				if conf.Ucfirst {
					name = strings.Title(name)
				}
			}
			str.WriteString(name)
		}
	}
	return nil
}

func (gen *Generator) parseRecursive(node *Node, options *GenOptions, parseOptions *ParseOptions) (err error) {
	str, exp := options.Str, options.Exp
	noObjectIndex, parseConf, captures := parseOptions.NoObjectIndex, parseOptions.Conf, parseOptions.Captures
	curType := node.Type
	conf := gen.Conf
	isNot := node.Operator == "!"
	if isNot {
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
						str.WriteString("\"")
						str.WriteString(text)
						str.WriteString("\" ")
					}
					express := string(runes[pos.StartIndex+1 : pos.EndIndex-1])
					ast, _ := exp.Parse(express)
					var inner string
					if inner, err = gen.Build(ast, options, parseOptions); err != nil {
						return err
					}
					str.WriteString(inner)
					i = pos.EndIndex + 1
					if i >= total {
						break
					}
				}
				if i < total {
					str.WriteString(" \"")
					str.WriteString(string(runes[i-1 : total]))
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
			if err = gen.parseIdentifier(options, parseOptions, name, ExpName); err != nil {
				return err
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
									isStaticOk = false
								}
							} else if first == "capture" {
								keyName := "$fet.capture." + second
								if variable, ok := (*captures)[keyName]; ok {
									str.WriteString("template \"" + second + "\" " + variable)
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
						return err
					}
					isParsed = true
				}
			}
		}
		if !isParsed {
			addIndexFn()
			if err = gen.parseRecursive(root, options, parseOptions); err != nil {
				return err
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
							str.WriteString(strings.Title(prop))
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
								ident = strings.Title(ident)
							}
							str.WriteString("\"")
							str.WriteString(ident)
							str.WriteString("\"")
						} else {
							if err = gen.parseIdentifier(options, parseOptions, ident, ObjectField); err != nil {
								return err
							}
						}
					} else {
						if err = gen.parseRecursive(cur, options, parseOptions); err != nil {
							return err
						}
					}
				} else {
					if err = gen.parseRecursive(cur, options, parseOptions); err != nil {
						return err
					}
				}
			}
		}
		if !noObjectIndex && !isStatic {
			str.WriteString(")")
		} else {
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
					return err
				}
				if _, ok := NoNeedIndexFuncs[name]; ok {
					parseOptions.NoObjectIndex = true
				}
				isParsed = true
			}
		}
		if !isParsed {
			if err = gen.parseRecursive(root, options, parseOptions); err != nil {
				return err
			}
		}
		str.WriteString(SPACE)
		for i, total := 0, len(args); i < total; i++ {
			if i > 0 {
				str.WriteString(SPACE)
			}
			if err = gen.parseRecursive(args[i], options, parseOptions); err != nil {
				return err
			}
		}
		str.WriteString(")")
	} else {
		op := node.Operator
		if name, ok := operatorFnNames[op]; ok {
			str.WriteString("(")
			str.WriteString(name)
			str.WriteString(SPACE)
		}
		gen.wrapToFloat(node.Left, options, parseOptions, op)
		str.WriteString(SPACE)
		gen.wrapToFloat(node.Right, options, parseOptions, op)
		str.WriteString(")")
	}
	if isNot {
		str.WriteString(")")
	}
	return nil
}

// New for Generator
func New(conf *GenConf) *Generator {
	return &Generator{
		Conf: conf,
	}
}
