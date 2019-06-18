package fet

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fet/lib/expression"
	"fet/lib/generator"
	"fet/utils"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

// Type type
type Type int

// Runes type
type Runes []rune

// MatchTagFn func
type MatchTagFn func(strs *Runes, index int, total int) (int, bool)

// ValidateFn func
type ValidateFn func(node *Node) string

// Config struct
type Config struct {
	LeftDelimiter  string
	RightDelimiter string
	CommentSymbol  string
	TemplateDir    string
	CompileDir     string
	Ucfirst        bool
	Ignores        []string
}

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

// Indexs struct
type Indexs struct {
	StartIndex int
	EndIndex   int
}

// Quote struct
type Quote struct {
	Indexs
	Variables []*Indexs
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
	LocalNS      string
	Includes     *[]string
	Extends      *[]string
}

// Node struct
type Node struct {
	Parent       *Node
	Type         Type
	Name         string
	Content      string
	Pwd          string
	Props        Props
	IsClosed     bool
	Features     []*Node
	Childs       []*Node
	Current      *Node
	Quotes       []*Quote
	Context      *Runes
	GlobalScopes []string
	LocalScopes  []string
	Fet          *Fet
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
		LeftDelimiter:  "{{",
		RightDelimiter: "}}",
		CommentSymbol:  "*",
		TemplateDir:    "templates",
		CompileDir:     "templates_c",
		Ucfirst:        true,
		Ignores:        []string{"inc/*"},
	}
	supportTags = map[string]Type{
		"include": SingleType,
		"extends": SingleType,
		"for":     BlockStartType,
		"if":      BlockStartType,
		"elseif":  BlockFeatureType,
		"else":    BlockFeatureType,
		"block":   BlockStartType,
	}
	validateFns = map[string]ValidateFn{
		"if":      validIfTag,
		"else":    validElseTag,
		"elseif":  validElseifTag,
		"for":     validForTag,
		"block":   validBlockTag,
		"include": validIncludeTag,
		"extends": validExtendsTag,
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
	exp, gen := fet.exp, fet.gen
	includes, extends := options.Includes, options.Extends
	name, content := node.Name, node.Content
	ld, rd := fet.LeftDelimiter, fet.RightDelimiter
	parentScopes, localScopes := options.ParentScopes, options.LocalScopes
	parentNS, localNS := options.ParentNS, options.LocalNS
	copyScopes := (*localScopes)[:]
	currentScopes := append(copyScopes, node.GlobalScopes...)
	namespace := func(name string) (bool, string) {
		if contains(currentScopes, name) {
			return true, localNS
		} else if contains(parentScopes, name) {
			return true, parentNS
		}
		return false, ""
	}
	toError := func(err error) error {
		return node.halt(err.Error())
	}
	switch node.Type {
	case CommentType:
		// output nothing
	case TextType:
		result = strings.TrimSpace(content)
	case AssignType, OutputType:
		// fmt.Println("content", content)
		ast, expErr := exp.Parse(content)
		if expErr != nil {
			err = toError(expErr)
		} else {
			//
			code := gen.Build(ast, namespace)
			if node.Type == AssignType {
				result = ld + "$" + name + localNS + ":=" + code + rd
			} else {
				result = "{{" + code + "}}"
			}
		}
	case SingleType:
		if name == "include" {
			tpl := fet.getRealTplPath(node.Content, path.Join(node.Pwd, ".."))
			ctx := md5.New()
			ctx.Write([]byte(tpl))
			curNS := hex.EncodeToString(ctx.Sum(nil))
			if contains(*includes, tpl) || contains(*extends, tpl) {
				err = fmt.Errorf("the include file '%s' has a loop dependence", tpl)
			} else {
				incOptions := &CompileOptions{
					ParentNS:     localNS,
					LocalNS:      "_" + curNS,
					ParentScopes: currentScopes[:],
					LocalScopes:  &[]string{},
					Extends:      extends,
					Includes:     includes,
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
		if name == "for" {
			props := node.Props
			target := props["list"].Raw
			ast, expErr := exp.Parse(target)
			if expErr != nil {
				err = toError(expErr)
			} else {
				code := gen.Build(ast, namespace)
				result = "{{range $" + props["key"].Raw + ", $" + props["item"].Raw + " := " + code + "}}"
			}
		} else if name == "if" {
			ast, expErr := exp.Parse(content)
			if expErr != nil {
				err = toError(expErr)
			} else {
				code := gen.Build(ast, namespace)
				result = "{{if " + code + "}}"
			}
		}
	case BlockFeatureType:
		if name == "elseif" {
			ast, expErr := exp.Parse(content)
			if expErr != nil {
				err = toError(expErr)
			} else {
				code := gen.Build(ast, namespace)
				result = "{{else if " + code + "}}"
			}
		} else if name == "else" {
			result = "{{else}}"
		}
	case BlockEndType:
		if name == "block" {
			blockScopes := node.LocalScopes
			if len(blockScopes) > 0 {
				*options.LocalScopes = append(*options.LocalScopes, blockScopes...)
			}
		} else {
			result = "{{end}}"
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
func validIfTag(node *Node) (errmsg string) {
	if node.Content == "" {
		errmsg = "the \"if\" tag does not have a condition expression"
	}
	return
}
func validElseTag(node *Node) (errmsg string) {
	if node.Content != "" {
		errmsg = "the \"else\" tag should not use a condition expression"
	} else if errmsg = validIfPrevCondOk(node); errmsg == "" {
		errmsg = validIfBlockCorrect(node, "if")
	}
	return
}
func validElseifTag(node *Node) (errmsg string) {
	if node.Content == "" {
		errmsg = "the \"elseif\" tag does not have a condition expression"
	} else if errmsg = validIfPrevCondOk(node); errmsg == "" {
		errmsg = validIfBlockCorrect(node, "if")
	}
	return
}
func validBlockTag(node *Node) (errmsg string) {
	block := node.Parent
	if block != nil && block.Parent != nil {
		errmsg = "the \"block\" tag should be root tag,can not appears in \"" + block.Parent.Name + "\""
	} else {
		errmsg = validIfOnlyOneStrParam(node)
	}
	return
}
func validForTag(node *Node) (errmsg string) {
	content := node.Content
	runes := Runes(node.Content)
	hasKey := false
	isComma := false
	prevIsSpace := true
	prevIsComma := false
	segs := []string{}
	total := 0
	maxNum := 2
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
	if len(segs) < 3 {
		errmsg = "the \"" + content + "\" tag is not correct"
	} else {
		inIndex := 1
		if hasKey && segs[1] == "," {
			inIndex = 3
		}
		if segs[inIndex] != "in" {
			errmsg = "wrong syntax \"for\" block"
		} else {
			node.Props["item"] = &Prop{
				Raw: segs[0],
			}
			if inIndex == 1 {
				node.Props["key"] = &Prop{
					Raw: "_",
				}
			} else {
				node.Props["key"] = &Prop{
					Raw: segs[inIndex-1],
				}
			}
			node.Props["list"] = &Prop{
				Raw: segs[inIndex+1],
			}
		}
	}
	return
}
func validIncludeTag(node *Node) (errmsg string) {
	errmsg = validIfOnlyOneStrParam(node)
	return
}
func validExtendsTag(node *Node) (errmsg string) {
	if node.Parent != nil {
		errmsg = "the \"extends\" tag should be root tag,can not appears in \"" + node.Parent.Name + "\""
	} else {
		errmsg = validIfOnlyOneStrParam(node)
	}
	return
}

// Validate method for Node
func (node *Node) Validate() (errmsg string) {
	if node.Type == UnknownType {
		errmsg = "unknown type"
	} else if !node.IsClosed {
		errmsg = "the tag is not closed"
	} else {
		name := node.Name
		if fn, exists := validateFns[name]; exists {
			errmsg = fn(node)
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
}

func mergeConfig(options *Config) *Config {
	conf := &Config{}
	*conf = *defConfig
	if options.Ucfirst {
		conf.Ucfirst = options.Ucfirst
	}
	if options.TemplateDir != "" {
		conf.TemplateDir = options.TemplateDir
	}
	if options.CompileDir != "" {
		conf.CompileDir = options.CompileDir
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
		Ucfirst: config.Ucfirst,
	})
	exp := expression.New()
	cwd, err := os.Executable()
	if err != nil {
		cwd = ""
	} else {
		cwd = filepath.Dir(cwd)
	}
	fet = &Fet{
		Config: config,
		Params: params,
		gen:    gen,
		exp:    exp,
		datas:  make(map[string]interface{}),
		cwd:    cwd,
	}
	fet.compileDir = fet.getLastDir(config.CompileDir)
	fet.templateDir = fet.getLastDir(config.TemplateDir)
	fmt.Println(fet.compileDir, fet.templateDir)
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
func (fet *Fet) parse(codes string, pwd string) (result *NodeList, err error) {
	var (
		isInComment, isTagStart, isInBlockTag, isSubTemplate bool
		node                                                 *Node
		variable                                             *Indexs
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
	resetGlobals := func(block *Node) {
		prevFeature := block.Current
		locals := prevFeature.LocalScopes
		if locals != nil {
			globals = globals[:len(globals)-len(locals)]
		} else {
			globals = globals[:]
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
				} else if code == "`" {
					// variable
					if variable != nil {
						variable.EndIndex = i + 1
						quote.Variables = append(quote.Variables, variable)
						variable = nil
					} else {
						variable = &Indexs{
							StartIndex: i,
						}
					}
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
									resetGlobals(block)
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
						if utils.IsIdentifier(ident) {
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
								globals = globals[:]
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
						node.EndIndex = curIndex
						isUnknownType := node.Type == UnknownType
						if node.Type == BlockEndType || isUnknownType {
							name := rtrim(&strs, markIndex+1, i)
							block := getLastBlock()
							if block == nil {
								if isUnknownType {
									if name == "" {
										err = node.halt("empty tag")
									} else {
										node.Type = OutputType
										node.Content = name
										setFeatureChild(node)
									}
								} else {
									err = node.halt("wrong end tag \"%s\"", name)
								}
								break
							} else {
								if isUnknownType {
									if curType, exists := supportTags[name]; exists && curType == BlockFeatureType {
										node.Name = name
										resetGlobals(block)
										block.AddFeature(node)
										closeTag(node, curIndex)
									} else {
										node.Type = OutputType
										node.Content = name
										setFeatureChild(node)
									}
								} else {
									if block.Name != name {
										err = block.halt("the block tag \"%s\" is not closed", block.Name)
										break
									} else {
										node.Name = name
										closeTag(node, curIndex)
										closeTag(block, curIndex)
										if name == "block" {
											isInBlockTag = false
											current := block.Current
											node.LocalScopes = current.LocalScopes[:]
											node.Content = current.Content
											current.Childs = queues[blockStartIndex:len(queues)]
										}
										resetGlobals(block)
										blocks = blocks[:len(blocks)-1]
									}
								}
							}
						} else {
							// validate tag
							noSpaceIndex := ltrimIndex(&strs, markIndex+1, i)
							node.Content = rtrim(&strs, noSpaceIndex, i)
							if errmsg := node.Validate(); errmsg != "" {
								err = node.halt(errmsg)
								break
							}
							if node.Type == BlockStartType && node.Name == "for" && isNeedScope() {
								props := node.Props
								itemProp := props["item"]
								keyProp := props["key"]
								globals = globals[:]
								globals = append(globals, itemProp.Raw)
								node.LocalScopes = append(node.LocalScopes, itemProp.Raw)
								if keyProp.Raw != "_" {
									globals = append(globals, keyProp.Raw)
									node.LocalScopes = append(node.LocalScopes, keyProp.Raw)
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
				globals := globals[:]
				node = &Node{
					Indexs: Indexs{
						StartIndex: i,
					},
					Position:     position,
					Props:        Props{},
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
					// if next == "=" || next == "-" {
					// 	// output
					// 	node.Type = OutputType
					// 	curIndex = nextIndex
					// } else
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
			globals := globals[:]
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
	if node != nil && !node.IsClosed {
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
func (fet *Fet) Display(tplpath string) {

}

// Assign method
func (fet *Fet) Assign(name string, value interface{}) {

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
						index += count
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
	if first == '/' {
		return dir
	} else if first == '.' {
		return path.Join(fet.cwd, dir)
	} else {
		return dir
	}
}

// Compile file
func (fet *Fet) Compile(tpl string) error {
	var (
		result string
		err    error
	)
	parentScopes := []string{}
	localScopes := []string{}
	includes := []string{}
	extends := []string{}
	tplFile := fet.getRealTplPath(tpl, fet.templateDir)
	compileFile := fet.getRealTplPath(tpl, fet.compileDir)
	options := &CompileOptions{
		ParentScopes: parentScopes,
		LocalScopes:  &localScopes,
		Includes:     &includes,
		Extends:      &extends,
	}
	defer func() {
		if err != nil {
			fmt.Println("compile fail:", err.Error())
		} else {
			fmt.Println("compile success")
		}
	}()
	fmt.Println("compile file:", tplFile, "--->", compileFile)
	if result, err = fet.compileFileContent(tplFile, options); err != nil {
		return err
	}
	dir := path.Dir(compileFile)
	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
		}
		if err != nil {
			return fmt.Errorf("can not open the compile dir:" + dir)
		}
	}
	err = ioutil.WriteFile(compileFile, []byte(result), 0644)
	if err != nil {
		return fmt.Errorf("compile file '" + compileFile + "' failure:" + err.Error())
	}
	return nil
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
		return fet.Compile(files[0])
	}
	var (
		wg   sync.WaitGroup
		errs []string
	)
	wg.Add(total)
	go func() {
		for _, tpl := range files {
			err := fet.Compile(tpl)
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
