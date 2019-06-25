package generator

import (
	"fmt"
	"strconv"
	"strings"

	e "github.com/fefit/fet/lib/expression"
	t "github.com/fefit/fet/types"
)

// Node for type alias
type Node = e.Node

// GenConf of generator
type GenConf struct {
	Ucfirst bool
}

// Generator for parse code
type Generator struct {
	Conf *GenConf
}

type opFnNames map[string]string

const (
	// SPACE CONSTANT
	SPACE     = " "
	toFloatFn = "INJECT_TO_FLOAT"
	concatFn  = "concat"
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
func (gen *Generator) Build(node *Node, nsFn t.NamespaceFn, exp *e.Expression) string {
	// conf := gen.Conf
	var str strings.Builder
	gen.parseRecursive(node, nsFn, &str, exp, false)
	return str.String()
}

func (gen *Generator) wrapToFloat(node *Node, nsFn t.NamespaceFn, str *strings.Builder, exp *e.Expression, isNative bool) {
	if isNative {
		str.WriteString("(")
		str.WriteString(toFloatFn)
		str.WriteString(SPACE)
	}
	gen.parseRecursive(node, nsFn, str, exp, false)
	if isNative {
		str.WriteString(")")
	}
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
func (gen *Generator) parseIdentifier(name string, nsFn t.NamespaceFn, str *strings.Builder, fieldType FieldType) {
	conf := gen.Conf
	if val, ok := LiteralSymbols[name]; ok {
		if fieldType != ExpName {
			panic(fmt.Sprint("syntax error: unexpect token ", name))
		} else {
			str.WriteString(val)
		}
	} else {
		isVar, name := nsFn(name)
		if isVar {
			str.WriteString(name)
		} else {
			if fieldType != FuncName {
				str.WriteString(".")
				if conf.Ucfirst {
					name = strings.Title(name)
				}
			}
			str.WriteString(name)
		}
	}
}

func (gen *Generator) parseRecursive(node *Node, nsFn t.NamespaceFn, str *strings.Builder, exp *e.Expression, noObjectIndex bool) {
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
					str.WriteString(gen.Build(ast, nsFn, exp))
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
			gen.parseIdentifier(name, nsFn, str, ExpName)
		}
	} else if curType == "object" {
		args := node.Arguments
		total := len(args)
		if !noObjectIndex {
			str.WriteString("(index ")
		}
		root := node.Root
		isParsed := false
		if root.Type == "raw" {
			if t, ok := root.Token.(*e.IdentifierToken); ok {
				gen.parseIdentifier(string(t.Stat.Values), nsFn, str, ObjectRoot)
				isParsed = true
			}
		}
		if !isParsed {
			gen.parseRecursive(root, nsFn, str, exp, noObjectIndex)
		}
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
						gen.parseIdentifier(ident, nsFn, str, ObjectField)
					}
				} else {
					gen.parseRecursive(cur, nsFn, str, exp, noObjectIndex)
				}
			} else {
				gen.parseRecursive(cur, nsFn, str, exp, noObjectIndex)
			}
		}
		if !noObjectIndex {
			str.WriteString(")")
		}
	} else if curType == "function" {
		root := node.Root
		args := node.Arguments
		str.WriteString("(")
		isParsed := false
		if root.Type == "raw" {
			if t, ok := root.Token.(*e.IdentifierToken); ok {
				name := string(t.Stat.Values)
				gen.parseIdentifier(name, nsFn, str, FuncName)
				if _, ok := NoNeedIndexFuncs[name]; ok {
					noObjectIndex = true
				}
				isParsed = true
			}
		}
		if !isParsed {
			gen.parseRecursive(root, nsFn, str, exp, noObjectIndex)
		}
		str.WriteString(SPACE)
		for i, total := 0, len(args); i < total; i++ {
			if i > 0 {
				str.WriteString(SPACE)
			}
			gen.parseRecursive(args[i], nsFn, str, exp, noObjectIndex)
		}
		str.WriteString(")")
	} else {
		op := node.Operator
		isNativeCompare := false
		if name, ok := operatorFnNames[op]; ok {
			if _, ok := compareFnNames[op]; ok {
				isNativeCompare = true
			}
			str.WriteString("(")
			str.WriteString(name)
			str.WriteString(SPACE)
		}
		gen.wrapToFloat(node.Left, nsFn, str, exp, isNativeCompare)
		str.WriteString(SPACE)
		gen.wrapToFloat(node.Right, nsFn, str, exp, isNativeCompare)
		str.WriteString(")")
	}
	if isNot {
		str.WriteString(")")
	}
}

// New for Generator
func New(conf *GenConf) *Generator {
	return &Generator{
		Conf: conf,
	}
}
