package generator

import (
	"strconv"
	"strings"

	exp "github.com/fefit/fet/lib/expression"
	t "github.com/fefit/fet/types"
)

type Node = exp.Node
type GenConf struct {
	Ucfirst bool
}
type Generator struct {
	Conf *GenConf
}

var (
	operatorFnNames = map[string]string{
		"||":      "or",
		"&&":      "and",
		"!":       "not",
		">=":      "ge",
		">":       "gt",
		"==":      "eq",
		"<":       "lt",
		"<=":      "le",
		"!=":      "ne",
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
	literalSymbols = map[string]string{
		"true":  "true",
		"false": "false",
		"null":  "nil",
		"nil":   "nil",
	}
)

func (gen *Generator) Build(node *Node, nsFn t.NamespaceFn) string {
	// conf := gen.Conf
	var str strings.Builder
	gen.parseRecursive(node, nsFn, false, &str)
	return str.String()
}

func (gen *Generator) parseRecursive(node *Node, nsFn t.NamespaceFn, isObject bool, str *strings.Builder) {
	curType := node.Type
	conf := gen.Conf
	space := " "
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
			if val, ok := literalSymbols["name"]; ok {
				str.WriteString(val)
			} else {
				if isObject {
					isVar, suffix := nsFn(name)
					if isVar {
						str.WriteString("$")
						str.WriteString(name)
						str.WriteString(suffix)
					} else {
						if conf.Ucfirst {
							name = strings.Title(name)
						}
						str.WriteString(".")
						str.WriteString(name)
					}
				} else {
					str.WriteString(name)
				}
			}
		}
	} else if curType == "object" {
		args := node.Arguments
		total := len(args)
		str.WriteString("(index ")
		gen.parseRecursive(node.Root, nsFn, true, str)
		str.WriteString(space)
		for i := 0; i < total; i++ {
			cur := args[i]
			curType := cur.Type
			str.WriteString(space)
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
				} else if t, ok := token.(*exp.IdentifierToken); ok && cur.Operator == "." {
					ident := string(t.Stat.Values)
					if conf.Ucfirst {
						ident = strings.Title(ident)
					}
					str.WriteString("\"")
					str.WriteString(ident)
					str.WriteString("\"")
				} else {
					gen.parseRecursive(cur, nsFn, isObject, str)
				}
			} else {
				gen.parseRecursive(cur, nsFn, isObject, str)
			}
		}
		str.WriteString(")")
	} else if curType == "function" {
		root := node.Root
		args := node.Arguments
		str.WriteString("(")
		gen.parseRecursive(root, nsFn, isObject, str)
		str.WriteString(space)
		for i, total := 0, len(args); i < total; i++ {
			if i > 0 {
				str.WriteString(space)
			}
			gen.parseRecursive(args[i], nsFn, isObject, str)
		}
		str.WriteString(")")
	} else {
		op := node.Operator
		if name, ok := operatorFnNames[op]; ok {
			str.WriteString("(")
			str.WriteString(name)
			str.WriteString(space)
		}
		gen.parseRecursive(node.Left, nsFn, isObject, str)
		str.WriteString(space)
		gen.parseRecursive(node.Right, nsFn, isObject, str)
		str.WriteString(")")
	}
}

func New(conf *GenConf) *Generator {
	return &Generator{
		Conf: conf,
	}
}
