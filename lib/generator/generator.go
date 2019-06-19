package generator

import (
	"fmt"
	"strconv"
	"strings"

	exp "github.com/fefit/fet/lib/expression"
	t "github.com/fefit/fet/types"
)

// Node for type alias
type Node = exp.Node

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
	// SPACE FOR
	SPACE     = " "
	toFloatFn = "INJECT_TO_FLOAT"
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
			"||":      "or",
			"&&":      "and",
			"!":       "not",
			"+":       "INJECT_PLUS",
			"-":       "INJECT_MINUS",
			"*":       "INJECT_MULTIPLE",
			"/":       "INJECT_DIVIDE",
			"%":       "INJECT_MOD",
			"&":       "INJECT_BITAND",
			" bitor ": "INJECT_BITOR",
			"^":       "INJECT_BITXOR",
			"**":      "INJECT_POWER",
		}
		for key, name := range compareFnNames {
			ops[key] = name
		}
		return ops
	}()
	literalSymbols = map[string]string{
		"true":  "true",
		"false": "false",
		"null":  "nil",
		"nil":   "nil",
	}
)

// Build for code
func (gen *Generator) Build(node *Node, nsFn t.NamespaceFn) string {
	// conf := gen.Conf
	var str strings.Builder
	gen.parseRecursive(node, nsFn, &str)
	return str.String()
}

func (gen *Generator) wrapToFloat(node *Node, nsFn t.NamespaceFn, str *strings.Builder, isNative bool) {
	if isNative {
		str.WriteString("(")
		str.WriteString(toFloatFn)
		str.WriteString(SPACE)
	}
	gen.parseRecursive(node, nsFn, str)
	if isNative {
		str.WriteString(")")
	}
}

// parse identifier
func (gen *Generator) parseIdentifier(name string, nsFn t.NamespaceFn, str *strings.Builder, isSubField bool) {
	conf := gen.Conf
	if val, ok := literalSymbols[name]; ok {
		if isSubField {
			panic(fmt.Sprint("synatax error: unexpect token ", name))
		} else {
			str.WriteString(val)
		}
	} else {
		isVar, suffix := nsFn(name)
		if isVar {
			str.WriteString("$")
			str.WriteString(name)
			str.WriteString(suffix)
		} else {
			if conf.Ucfirst {
				name = strings.Title(name)
			}
			if !isSubField {
				str.WriteString(".")
			}
			str.WriteString(name)
		}
	}
}

func (gen *Generator) parseRecursive(node *Node, nsFn t.NamespaceFn, str *strings.Builder) {
	curType := node.Type
	conf := gen.Conf
	if curType == "raw" {
		token := node.Token
		stat := token.GetStat()
		switch t := token.(type) {
		case *exp.StringToken:
			str.WriteString(string(stat.Values))
		case *exp.NumberToken:
			str.WriteString(strconv.FormatFloat(t.ToNumber(), 'f', -1, 64))
		case *exp.IdentifierToken:
			name := string(stat.Values)
			gen.parseIdentifier(name, nsFn, str, false)
		}
	} else if curType == "object" {
		args := node.Arguments
		total := len(args)
		str.WriteString("(index ")
		gen.parseRecursive(node.Root, nsFn, str)
		str.WriteString(SPACE)
		for i := 0; i < total; i++ {
			cur := args[i]
			curType := cur.Type
			str.WriteString(SPACE)
			if curType == "raw" {
				token := cur.Token
				if t, ok := token.(*exp.StringToken); ok {
					prop := string(t.Stat.Values)
					if conf.Ucfirst {
						str.WriteString(strings.Title(prop))
					} else {
						str.WriteString(prop)
					}
				} else if t, ok := token.(*exp.NumberToken); ok {
					index := t.ToNumber()
					str.WriteString(strconv.FormatInt(int64(index), 10))
				} else if t, ok := token.(*exp.IdentifierToken); ok {
					ident := string(t.Stat.Values)
					if cur.Operator == "." {
						if conf.Ucfirst {
							ident = strings.Title(ident)
						}
						str.WriteString("\"")
						str.WriteString(ident)
						str.WriteString("\"")
					} else {
						gen.parseIdentifier(ident, nsFn, str, true)
					}
				} else {
					gen.parseRecursive(cur, nsFn, str)
				}
			} else {
				gen.parseRecursive(cur, nsFn, str)
			}
		}
		str.WriteString(")")
	} else if curType == "function" {
		root := node.Root
		args := node.Arguments
		str.WriteString("(")
		gen.parseRecursive(root, nsFn, str)
		str.WriteString(SPACE)
		for i, total := 0, len(args); i < total; i++ {
			if i > 0 {
				str.WriteString(SPACE)
			}
			gen.parseRecursive(args[i], nsFn, str)
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
		gen.wrapToFloat(node.Left, nsFn, str, isNativeCompare)
		str.WriteString(SPACE)
		gen.wrapToFloat(node.Right, nsFn, str, isNativeCompare)
		str.WriteString(")")
	}
}

// New for Generator
func New(conf *GenConf) *Generator {
	return &Generator{
		Conf: conf,
	}
}
