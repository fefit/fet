package expression

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"unicode"

	"github.com/fefit/fet/types"
)

// Indexs
type Indexs = types.Indexs

// Type token
type Type int

// Runes rune slice
type Runes []rune

// Flags map
type Flags map[string]bool

// Operator string
type Operator struct {
	Runes    Runes
	Priority int
}

// Expression struct
type Expression struct {
	Parser  *Parser
	Initial bool
}

// Node struct
type Node struct {
	Left      *Node
	Right     *Node
	Root      *Node
	Token     AnyToken
	Type      string
	Operator  string
	Arguments []*Node
}

// OperatorList data
type OperatorList struct {
	Keys   []string
	Values map[string]*Operator
}

// Quote rune
const (
	Quote              = '"'
	LeftRoundBracket   = '('
	RightRoundBracket  = ')'
	LeftSquareBracket  = '['
	RightSquareBracket = ']'
	Minus              = '-'
	Plus               = '+'
	Underline          = '_'
	Translate          = '\\'
	Space              = ' '
	Bitor              = "bitor"
	Dollar             = '$'
	VarSymbol          = '`'
)

var (
	operators = [][]string{
		[]string{","},
		[]string{"||"},
		[]string{"&&"},
		[]string{Bitor},
		[]string{"^"},
		[]string{"&"},
		[]string{"==", "!="},
		[]string{">=", "<=", "<", ">"},
		[]string{"+", "-"},
		[]string{"*", "/", "%"},
		[]string{"**"},
		[]string{"!"},
		[]string{".", "|", ":", "(", ")", "[", "]"},
	}
	keywordOperators = map[string]string{
		"bitor": "bitor",
		"and":   "&&",
		"or":    "||",
		"not":   "!",
		"eq":    "==",
		"ne":    "!=",
		"lt":    "<",
		"le":    "<=",
		"gt":    ">",
		"ge":    ">=",
	}
	operatorList = func() OperatorList {
		preIndex := 8     // add +- first
		ignoreIndex := 12 // ignore () []
		keys, values := []string{}, map[string]*Operator{}
		keys = append(keys, operators[preIndex]...)
		for i, total := 0, len(operators); i < total; i++ {
			ops := operators[i]
			for j, count := 0, len(ops); j < count; j++ {
				key := ops[j]
				values[key] = &Operator{
					Priority: i,
					Runes:    Runes(key),
				}
				if i == preIndex || (i == ignoreIndex && j > 2) {
					continue
				}
				keys = append(keys, key)
			}
		}
		return OperatorList{
			Keys:   keys,
			Values: values,
		}
	}()
	bases = map[rune]int{
		'b': 2,
		'o': 8,
		'x': 16,
	}
)

// Any interface
type Any interface{}

// AnyToken make sure is token type
type AnyToken interface {
	Add(s rune) (bool, bool, bool, error)
	Validate(prevTokens []AnyToken) (AnyToken, error)
	SetStat(stat *TokenStat)
	GetStat() *TokenStat
}

func typeOf(token Any) string {
	name := fmt.Sprintf("%T", token)
	rns := Runes(name)
	total := len(rns)
	for total >= 0 {
		total--
		if r := rns[total]; r == '.' {
			break
		}
	}
	return string(rns[total+1:])
}

func isBracket(token AnyToken) bool {
	switch token.(type) {
	case *LeftBracketToken, *RightBracketToken, *LeftSquareBracketToken, *RightSquareBracketToken:
		return true
	default:
		return false
	}
}

func getNoSpaceTokens(tokens []AnyToken, num int) (isPrevSpace bool, result []AnyToken) {
	isTwo := num == 2
	if num != 1 && !isTwo {
		log.Fatal("wrong num argments:getNoSpaceTokens")
	}
	prevs := getPrevTokens(tokens, num*2)
	p0, p1 := prevs[0], prevs[1]
	val, match := p0, p1
	if _, isSpace := p0.(*SpaceToken); isSpace {
		val = p1
		isPrevSpace = true
		if isTwo {
			p2, p3 := prevs[2], prevs[3]
			if _, isStillSpace := p2.(*SpaceToken); isStillSpace {
				match = p3
			} else {
				match = p2
			}
		}
	}
	result = append(result, val)
	if isTwo {
		result = append(result, match)
	}
	return
}

// TokenStat struct
type TokenStat struct {
	StartIndex int
	Index      int
	ParseIndex int
	Logics     Flags
	Context    *Runes
	Values     Runes
	RBLevel    int
	RBSubLevel int
	SBLevel    int
	SBSubLevel int
}

// Token struct
type Token struct {
	IsComplete bool
	IsBegin    bool
	Stat       *TokenStat
}

// ValidateNextFn func
type ValidateNextFn func(token AnyToken) error

// Validate for token
func getPrevTokens(tokens []AnyToken, num int) []AnyToken {
	i, count, prevTokens := 1, len(tokens), []AnyToken{}
	for i <= num {
		if count >= i {
			prevTokens = append(prevTokens, tokens[count-i])
		} else {
			prevTokens = append(prevTokens, nil)
		}
		i++
	}
	return prevTokens
}

// GetStat for token
func (token *Token) GetStat() *TokenStat {
	return token.Stat
}

// Validate for token
func (token *Token) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	return
}

// Judge for token
func (token *Token) Judge() (ok bool, isComplete bool, retry bool, err error) {
	if token.IsBegin {
		token.IsComplete = true
		isComplete = true
	}
	return
}

// AddBracket for token
func (token *Token) AddBracket(s rune, bracket rune) (ok bool, isComplete bool, retry bool, err error) {
	if s == bracket {
		ok = true
		isComplete = true
		stat := token.Stat
		stat.Values = append(stat.Values, bracket)
		// token.Values = append(token.Values, bracket)
	}
	return
}

// SetStat for token
func (token *Token) SetStat(stat *TokenStat) {
	token.Stat = stat
}

// LeftBracketToken struct
type LeftBracketToken struct {
	Token
}

// Add for LeftBracketToken
func (bracket *LeftBracketToken) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	return bracket.AddBracket(s, LeftRoundBracket)
}

// Validate for LeftBracketToken
func (bracket *LeftBracketToken) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	hasSpace, prevs := getNoSpaceTokens(tokens, 1)
	prev := prevs[0]
	if prev == nil {
		return
	}
	switch prev.(type) {
	case *OperatorToken, *LeftBracketToken:
	case *IdentifierToken, *RightSquareBracketToken:
		if hasSpace {
			return nil, fmt.Errorf("wrong space between function name and (")
		}
		stat := bracket.Stat
		stat.Logics = Flags{
			"IsFunc": true,
		}
	default:
		return nil, fmt.Errorf("wrong left round bracket previous token")
	}
	return
}

// RightBracketToken struct
type RightBracketToken struct {
	Token
}

// Add for SpaceToken
func (bracket *RightBracketToken) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	return bracket.AddBracket(s, RightRoundBracket)
}

// Validate for RightBracketToken
func (bracket *RightBracketToken) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	stat := bracket.Stat
	if stat.RBLevel < 1 {
		return nil, fmt.Errorf("wrong right round bracket without start:%d", stat.StartIndex)
	}
	prevs := getPrevTokens(tokens, 1)
	p0 := prevs[0]
	if p0, ok := p0.(*LeftBracketToken); ok && !p0.Stat.Logics["IsFunc"] {
		return nil, fmt.Errorf("empty expression")
	}
	return
}

// LeftSquareBracketToken struct
type LeftSquareBracketToken struct {
	Token
}

// Add for LeftSquareBracketToken
func (bracket *LeftSquareBracketToken) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	return bracket.AddBracket(s, LeftSquareBracket)
}

// Validate for LeftSquareBracketToken
func (bracket *LeftSquareBracketToken) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	prevs := getPrevTokens(tokens, 1)
	p0 := prevs[0]
	switch p0.(type) {
	case *IdentifierToken, *RightSquareBracketToken, *RightBracketToken:
	default:
		return nil, fmt.Errorf("wrong member or index of object")
	}
	return
}

// RightSquareBracketToken struct
type RightSquareBracketToken struct {
	Token
}

// Add for SpaceToken
func (bracket *RightSquareBracketToken) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	return bracket.AddBracket(s, RightSquareBracket)
}

// Validate for RightSquareBracketToken
func (bracket *RightSquareBracketToken) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	stat := bracket.Stat
	if stat.SBLevel < 0 {
		return nil, fmt.Errorf("can not find matched left square bracket:%d", stat.StartIndex)
	}
	_, prevs := getNoSpaceTokens(tokens, 2)
	val, match := prevs[0], prevs[1]
	if _, isLS := match.(*LeftSquareBracketToken); isLS {
		switch token := val.(type) {
		case *StringToken, *IdentifierToken:
		case *NumberToken:
			// will validate in number token
			logics := token.Stat.Logics
			if logics["IsFloat"] || logics["IsPower"] {
				return nil, fmt.Errorf("can't use float or power number as index")
			}
		default:
			return nil, fmt.Errorf("wrong member visit")
		}
	} else if _, isLS := val.(*LeftSquareBracketToken); isLS {
		return nil, fmt.Errorf("empty member")
	}
	return
}

// SpaceToken spaces
type SpaceToken struct {
	Token
}

// Add for SpaceToken
func (space *SpaceToken) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	if unicode.IsSpace(s) {
		ok = true
		if !space.IsBegin {
			space.IsBegin = true
		}
		// space.Values = append(space.Values, s)
		stat := space.Stat
		stat.Values = append(stat.Values, s)
		return
	}
	return space.Judge()
}

// IdentifierToken identifiers
type IdentifierToken struct {
	Token
}

// Add for identifiers
func (identifier *IdentifierToken) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	stat := identifier.Stat
	if isDigit := unicode.IsDigit(s); unicode.IsLetter(s) || isDigit || s == Underline || s == Dollar {
		if !identifier.IsBegin {
			if isDigit {
				return
			}
			identifier.IsBegin = true
		} else if s == Dollar {
			err = fmt.Errorf("the $ can only use in identifier head")
			return
		}
		stat.Values = append(stat.Values, s)
		ok = true
		return
	}
	vals := stat.Values
	if len(vals) == 1 && (vals[0] == Underline || vals[0] == Dollar) {
		err = fmt.Errorf("can not use single %s as identifier", string(vals[0]))
		return
	}
	return identifier.Judge()
}

// Validate for IdentifierToken
func (identifier *IdentifierToken) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	hasSpace, prevs := getNoSpaceTokens(tokens, 1)
	prev := prevs[0]
	stat := identifier.Stat
	ident := string(stat.Values)
	if op, ok := keywordOperators[ident]; ok {
		if op == "!" && prev == nil {
			// "not"
		} else if !hasSpace {
			return nil, fmt.Errorf("the keyword operator '%s' must have a left space after string or number", ident)
		}
		return buildOperatorToken(stat, op, ident), nil
	}
	if prev == nil {
		return
	}
	switch prev.(type) {
	case *OperatorToken, *LeftBracketToken, *LeftSquareBracketToken:
	default:
		return nil, fmt.Errorf("wrong idenfier token")
	}
	return
}

// StringToken strings
type StringToken struct {
	Token
	Variables []*Indexs
}

// Add for string
func (str *StringToken) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	stat := str.Stat
	if !str.IsBegin {
		if s == Quote {
			ok = true
			stat.Values = append(stat.Values, s)
			stat.Logics = Flags{}
			str.IsBegin = true
			return
		}
		return
	}
	ok = true
	logics := stat.Logics
	stat.Values = append(stat.Values, s)
	if logics["IsInTransalte"] {
		logics["IsInTransalte"] = false
		return
	}
	if s == Translate {
		logics["IsInTransalte"] = true
		logics["IsInVar"] = false
		return
	}
	if s == VarSymbol {
		if logics["IsInVar"] {
			last := str.Variables[len(str.Variables)-1]
			last.EndIndex = len(stat.Values)
		} else {
			logics["IsInVar"] = true
			str.Variables = append(str.Variables, &Indexs{
				StartIndex: len(stat.Values) - 1,
			})
		}
	}
	if s == Quote {
		isComplete = true
		str.IsComplete = true
		vars := str.Variables
		realVars := []*Indexs{}
		if len(vars) > 0 {
			for _, pos := range vars {
				if pos.EndIndex > pos.StartIndex+1 {
					realVars = append(realVars, pos)
				}
			}
			str.Variables = realVars
		}
	}
	return
}

// Validate for StringToken
func (str *StringToken) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	_, prevs := getNoSpaceTokens(tokens, 1)
	prev := prevs[0]
	if prev == nil {
		return
	}
	switch token := prev.(type) {
	case *LeftSquareBracketToken, *LeftBracketToken:
	case *OperatorToken:
		name := token.Name
		ops := operatorList.Values
		if value, exists := ops[name]; exists {
			name = string(value.Runes)
			if name == "+" || name == "!=" || name == "==" || name == ":" || name == "&&" || name == "||" {
				return
			}
		}
		return nil, fmt.Errorf("can not use operator %v with strings", name)
	default:
		return nil, fmt.Errorf("can not use string")
	}
	return
}

// NumberToken numbers
type NumberToken struct {
	Token
	Base     int
	Dicimals Runes
	Power    *NumberToken
}

func (number *NumberToken) setNumberBegin(s rune) {
	stat := number.Stat
	logics := stat.Logics
	logics["IsNumberBegin"] = true
	if s == '0' {
		logics["IsZeroBegin"] = true
	} else {
		stat.Values = append(stat.Values, s)
	}
}

// Add for string
func (number *NumberToken) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	if !number.IsBegin {
		if isDigit := unicode.IsDigit(s); isDigit || s == Plus || s == Minus {
			number.IsBegin = true
			ok = true
			stat := number.Stat
			stat.Logics = Flags{}
			if isDigit {
				number.setNumberBegin(s)
				return
			}
			if s == Plus {
				stat.Logics["IsPlus"] = true
			} else {
				stat.Logics["IsMinus"] = true
			}
		}
		return
	}
	stat := number.Stat
	logics := stat.Logics
	// has prefix (+-)
	if !logics["IsNumberBegin"] {
		if !unicode.IsDigit(s) {
			ok = false
			retry = true
			err = fmt.Errorf("should try a operator")
			return
		}
		ok = true
		number.setNumberBegin(s)
		return
	}
	// decimal
	if logics["IsBase"] {
		switch number.Base {
		case 2:
			ok = s == '0' || s == '1'
		case 8:
			ok = s >= '0' && s <= '7'
		case 16:
			ok = (s >= '0' && s <= '9') || (s >= 'a' && s <= 'f') || (s >= 'A' && s <= 'F')
		}
		if ok {
			stat.Values = append(stat.Values, s)
			return
		}
		if stat.Values != nil {
			isComplete = true
		} else {
			err = fmt.Errorf("unrecognize base %d:%v", number.Base, string(s))
		}
		return
	}
	// scientific notation
	if number.Power != nil {
		power := number.Power
		if powerOk, isPowerComplete, _, powerErr := power.Add(s); powerErr != nil {
			err = powerErr
		} else {
			if powerOk {
				powerLogics := power.Stat.Logics
				if powerLogics["IsFloat"] || powerLogics["IsBase"] || powerLogics["IsPower"] {
					err = fmt.Errorf("wrong scientific notation:%s", string(s))
				} else {
					ok = true
				}
			} else {
				if isPowerComplete {
					isComplete = true
					number.IsComplete = true
				} else {
					err = fmt.Errorf("wrong scientific notation: %v", s)
				}
			}
		}
		return
	}
	// float or integer, first judge if float
	isFloat := logics["IsFloat"]
	isStillInt := !isFloat && stat.Values != nil
	if isFloat || isStillInt {
		if isDigit := unicode.IsDigit(s); isDigit {
			ok = true
			if isFloat {
				number.Dicimals = append(number.Dicimals, s)
			} else {
				stat.Values = append(stat.Values, s)
			}
			return
		}
		if s == 'e' {
			if isFloat && number.Dicimals == nil {
				err = fmt.Errorf("wrong float without dicimals")
				return
			}
			ok = true
			logics["IsPower"] = true
			number.Power = &NumberToken{
				Token: Token{
					Stat: &TokenStat{},
				},
			}
			return
		}
		if s == '.' && isStillInt {
			ok = true
			logics["IsFloat"] = true
			return
		}
		ok = false
		isComplete = true
		return
	}
	// begin with 0, maybe float,decimals
	if logics["IsZeroBegin"] {
		if base, exists := bases[s]; exists {
			number.Base = base
			logics["IsBase"] = true
			ok = true
			return
		} else if s == '.' {
			ok = true
			logics["IsFloat"] = true
			return
		}
		// just zero
		isComplete = true
		return
	}
	return
}

func buildOperatorToken(stat *TokenStat, name string, keyword string) *OperatorToken {
	retryStat := &TokenStat{}
	*retryStat = *stat
	retryStat.Values = Runes(name)
	retryStat.Logics = Flags{}
	op := operatorList.Values[name]
	return &OperatorToken{
		Token: Token{
			IsBegin:    true,
			IsComplete: true,
			Stat:       retryStat,
		},
		Keyword: keyword,
		Exact:   op,
		Name:    name,
	}
}

// Validate for NumberToken
func (number *NumberToken) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	hasSpace, prevs := getNoSpaceTokens(tokens, 1)
	prev := prevs[0]
	if prev == nil {
		return
	}
	stat := number.Stat
	logics := stat.Logics
	switch token := prev.(type) {
	case *LeftSquareBracketToken, *LeftBracketToken:
	case *OperatorToken:
		name := token.Name
		if !hasSpace {
			if (name == "-" && logics["IsMinus"]) || (name == "+" && logics["IsPlus"]) {
				return nil, fmt.Errorf("can not use %#v%#v operator", name, name)
			}
		}
	default:
		if logics["IsMinus"] || logics["IsPlus"] {
			var name string
			if logics["IsMinus"] {
				name = "-"
				logics["IsMinus"] = false
			} else {
				name = "+"
				logics["IsPlus"] = false
			}
			retryToken := buildOperatorToken(stat, name, "")
			stat.StartIndex++
			stat.Index++
			return retryToken, nil
		}
		return nil, fmt.Errorf("wrong number")
	}
	return
}

// ToNumber for number token
func (number *NumberToken) ToNumber() float64 {
	stat := number.Stat
	logics := stat.Logics
	values := string(stat.Values)
	power := 1.0
	symbol := 1.0
	var num float64
	if logics["IsMinus"] {
		symbol = -1.0
	}
	if logics["IsBase"] {
		cur, _ := strconv.ParseInt(values, number.Base, 64)
		num = float64(cur)
	} else {
		num, _ = strconv.ParseFloat(values, 10)
		if number.Power != nil {
			power = number.Power.ToNumber()
		}
	}
	return symbol * num * power
}

// OperatorToken spaces
type OperatorToken struct {
	Token
	Maybes       []string
	Exact        *Operator
	Name         string
	CompareIndex int
	Keyword      string
}

// Validate for OperatorToken
func (op *OperatorToken) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	_, prevs := getNoSpaceTokens(tokens, 1)
	prev := prevs[0]
	name := op.Name
	switch prev.(type) {
	case *NumberToken, *IdentifierToken, *RightBracketToken, *RightSquareBracketToken:
	case *OperatorToken, nil:
		if name != "!" {
			return nil, fmt.Errorf("wrong operator token")
		}
	case *StringToken:
		if name == "&&" || name == "||" || name == "," || name == ":" || name == "|" {
			// allow operators
		} else {
			return nil, fmt.Errorf("wrong opeator with previous token 'string'")
		}
	default:
		return nil, fmt.Errorf("wrong operator token" + name)
	}
	return
}

// Add for OperatorToken
func (op *OperatorToken) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	target := operatorList.Keys
	all := operatorList.Values
	if op.IsBegin {
		target = op.Maybes
	}
	index := op.CompareIndex
	maybes := []string{}
	for _, key := range target {
		item := all[key]
		rn := item.Runes
		maxIndex := len(rn) - 1
		if index <= maxIndex && s == rn[index] {
			if index == maxIndex {
				op.Exact = item
				op.Name = key
				op.Stat.Values = rn
				ok = true
			} else {
				maybes = append(maybes, key)
			}
		}
	}
	if len(maybes) == 0 {
		if op.Exact != nil {
			isComplete = true
			op.IsComplete = true
		}
		return
	}
	if !op.IsBegin {
		op.IsBegin = true
	}
	ok = true
	op.Maybes = maybes
	op.CompareIndex++
	return
}

// EOFToken struct
type EOFToken struct {
	Token
}

// Add for EOF
func Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	err = fmt.Errorf("the eof token should not contains any character")
	return
}

// Validate for EOF
func (eof *EOFToken) Validate(tokens []AnyToken) (retryToken AnyToken, err error) {
	_, prevs := getNoSpaceTokens(tokens, 1)
	prev := prevs[0]
	switch prev.(type) {
	case *OperatorToken:
		return nil, fmt.Errorf("wrong operator token at last position")
	}
	return
}

// TokenNode for translate
type TokenNode struct {
	Token
	Node *Node
	Type string
}

// Add for nodetoken
func (node *TokenNode) Add(s rune) (ok bool, isComplete bool, retry bool, err error) {
	return
}

func setLevel(list map[int]int, level int) (subLevel int) {
	if value, exists := list[level]; exists {
		list[level] = value + 1
	} else {
		list[level] = 1
	}
	return list[level]
}

// Parser struct
type Parser struct {
	Current      AnyToken
	Tokens       []AnyToken
	RBLevel      int
	RBSubLevel   int
	SBLevel      int
	SBSubLevel   int
	BLList       map[int]int
	SLList       map[int]int
	CurBrSqLevel [2]int
	CurSqBrLevel [2]int
	Asserts      map[int]AnyToken
	TokenIndex   int
	Context      Runes
	IgnoreIndex  int
	TokenStat    *TokenStat
	NextMustBe   ValidateNextFn
	NextValids   int
}

// Init parserer
func (parser *Parser) Init() {
	tokens := []AnyToken{
		&StringToken{},
		&SpaceToken{},
		&IdentifierToken{},
		&NumberToken{},
		&OperatorToken{},
		&LeftBracketToken{},
		&RightBracketToken{},
		&LeftSquareBracketToken{},
		&RightSquareBracketToken{},
	}
	asserts := map[int]AnyToken{}
	for i, token := range tokens {
		asserts[i] = token
	}
	parser.Asserts = asserts
	parser.Reuse()
}

// Reuse for parser
func (parser *Parser) Reuse() {
	parser.RBLevel = 0
	parser.RBSubLevel = 0
	parser.SBLevel = 0
	parser.SBSubLevel = 0
	parser.BLList = map[int]int{}
	parser.SLList = map[int]int{}
	parser.CurBrSqLevel = [2]int{}
	parser.CurSqBrLevel = [2]int{}
	parser.TokenIndex = -1
	parser.IgnoreIndex = -1
	parser.TokenStat = &TokenStat{}
	parser.Context = nil
	parser.Tokens = nil
	parser.Current = nil
	parser.NextMustBe = nil
	parser.NextValids = 0
}

// Reset for parser
func (parser *Parser) Reset() {
	tokenPos := parser.TokenIndex
	asserts := parser.Asserts
	if value, exists := asserts[tokenPos]; exists {
		switch value.(type) {
		case *StringToken:
			asserts[tokenPos] = &StringToken{}
		case *SpaceToken:
			asserts[tokenPos] = &SpaceToken{}
		case *LeftBracketToken:
			asserts[tokenPos] = &LeftBracketToken{}
		case *RightBracketToken:
			asserts[tokenPos] = &RightBracketToken{}
		case *LeftSquareBracketToken:
			asserts[tokenPos] = &LeftSquareBracketToken{}
		case *RightSquareBracketToken:
			asserts[tokenPos] = &RightSquareBracketToken{}
		case *IdentifierToken:
			asserts[tokenPos] = &IdentifierToken{}
		case *NumberToken:
			asserts[tokenPos] = &NumberToken{}
		case *OperatorToken:
			asserts[tokenPos] = &OperatorToken{}
		default:
			panic("unexpect token")
		}
	}
	parser.TokenIndex = -1
}

// SetToken for parser
func (parser *Parser) SetToken(token AnyToken) {
	parser.Current = token
	stat := parser.TokenStat
	stat.StartIndex = len(parser.Context)
	stat.Index = len(parser.Tokens)
	stat.RBLevel = parser.RBLevel
	stat.RBSubLevel = parser.RBSubLevel
	stat.SBLevel = parser.SBLevel
	stat.SBSubLevel = parser.SBSubLevel
}

// Next for Parser
func (parser *Parser) Next() {
	if parser.Current != nil {
		parser.Tokens = append(parser.Tokens, parser.Current)
		parser.Current = nil
	}
	parser.TokenStat = &TokenStat{}
}

// Add for parser
func (parser *Parser) Add(s rune) error {
	var (
		ok, isComplete, retry bool
		err                   error
		retryToken            AnyToken
	)
	asserts := parser.Asserts
	current := parser.Current
	ignorePos := parser.IgnoreIndex
	currentStat := parser.TokenStat
	// types := []string{"string", "space", "leftBracket", "rightBracket", "leftSquare", "rightSquare", "identifier", "number", "operator"}
	defer func() {
		if ok && err == nil && !retry {
			parser.Context = append(parser.Context, s)
		}
	}()
	if current == nil {
		for i, total := 0, len(asserts); i < total; i++ {
			cur := asserts[i]
			// ignore pos
			if i <= ignorePos {
				if i == ignorePos {
					parser.IgnoreIndex = -1
				}
				continue
			}
			cur.SetStat(parser.TokenStat)
			if ok, isComplete, retry, err = cur.Add(s); ok {
				current = cur
				parser.TokenIndex = i
				parser.SetToken(cur)
				break
			}
		}
		if current == nil {
			return fmt.Errorf("can not find any token match ->" + string(s))
		}
	} else {
		ok, isComplete, retry, err = current.Add(s)
	}
	if err != nil {
		// only number token has prefix +- should retry
		if retry {
			context := parser.Context
			count := len(context)
			parser.IgnoreIndex = parser.TokenIndex
			parser.Reset()
			parser.Current = nil
			index := count - 1
			prefix := context[index]
			parser.Context = context[:index]
			parser.TokenStat = &TokenStat{}
			if err = parser.Add(prefix); err == nil {
				return parser.Add(s)
			}
		}
		return err
	}
	if isComplete {
		blList := parser.BLList
		slList := parser.SLList
		rbl := parser.RBLevel
		sbl := parser.SBLevel
		rbsl := parser.RBSubLevel
		sbsl := parser.SBSubLevel
		tokens := parser.Tokens
		switch current.(type) {
		case *LeftBracketToken:
			parser.RBSubLevel = setLevel(blList, rbl)
			currentStat.RBLevel = rbl + 1
			currentStat.RBSubLevel = parser.RBSubLevel
			parser.CurBrSqLevel = [2]int{sbl, sbsl}
			parser.RBLevel++
		case *RightBracketToken:
			levels := parser.CurBrSqLevel
			// e.g [(][)]
			if levels[0] != sbl || levels[1] != sbsl {
				return fmt.Errorf("wrong matched right bracket")
			}
			currentStat.RBLevel = rbl
			parser.RBLevel--
		case *LeftSquareBracketToken:
			currentStat.SBLevel = sbl + 1
			parser.SBSubLevel = setLevel(slList, sbl)
			currentStat.SBSubLevel = parser.SBSubLevel
			parser.CurSqBrLevel = [2]int{rbl, rbsl}
			parser.SBLevel++
		case *RightSquareBracketToken:
			levels := parser.CurSqBrLevel
			if levels[0] != rbl || levels[1] != rbsl {
				// e.g ([)(])
				return fmt.Errorf("wrong matched right square bracket")
			}
			currentStat.SBLevel = sbl
			parser.SBLevel--
		default:
		}
		if retryToken, err = current.Validate(tokens); err != nil {
			return err
		} else if retryToken != nil {
			// try number token into operator(+/-) and number
			// try identifier to keyword operators
			if _, err = retryToken.Validate(tokens); err != nil {
				return err
			}
			if _, isMP := current.(*NumberToken); isMP {
				// +-
				parser.Tokens = append(parser.Tokens, retryToken)
			} else {
				// keyword operator
				parser.Current = retryToken
				// next must has a space
				if token, ok := retryToken.(*OperatorToken); ok {
					var validNext ValidateNextFn
					validNext = func(t AnyToken) error {
						if _, isSpace := t.(*SpaceToken); isSpace {
							return nil
						}
						return fmt.Errorf("the keyword operator '%s' must have a right space", token.Keyword)
					}
					parser.NextMustBe = validNext
					parser.NextValids++
				}
			}
		}
		parser.Reset()
		// check if has validate next func
		if parser.NextMustBe != nil {
			valids := parser.NextValids
			if valids == 0 {
				if vErr := parser.NextMustBe(parser.Current); vErr != nil {
					return vErr
				}
				parser.NextMustBe = nil
			} else {
				// jump current
				parser.NextValids--
				if parser.NextValids > 0 {
					return fmt.Errorf("wrong token:%#v", parser.Current)
				}
			}
		}
		parser.Next()
		if !ok {
			return parser.Add(s)
		}
	}
	return nil
}

// New return Expression
func New() *Expression {
	parser := &Parser{}
	parser.Init()
	exp := &Expression{
		Parser: parser,
	}
	return exp
}

func bracketToOperator(name string, stat *TokenStat) *OperatorToken {
	operator := operatorList.Values[name]
	return &OperatorToken{
		Token: Token{
			IsBegin:    true,
			IsComplete: true,
			Stat:       stat,
		},
		Name:  name,
		Exact: operator,
	}
}

func getParsedNode(index int, parsed []AnyToken) *Node {
	token := parsed[index]
	if cur, ok := token.(*TokenNode); ok {
		return cur.Node
	}
	return &Node{
		Type:  "raw",
		Token: token,
	}
}

// parsed to ast
func (exp *Expression) toAst(tokens []AnyToken, isInFunc bool) (*TokenNode, error) {
	lastToken := Token{
		Stat: &TokenStat{},
	}
	if len(tokens) == 1 {
		token := tokens[0]
		return &TokenNode{
			Token: lastToken,
			Node: &Node{
				Type:  "raw",
				Token: token,
			},
		}, nil
	}
	var subs []AnyToken
	var levels [2]int
	isInBracket := false
	noRounds := []AnyToken{}
	// parse round brackets
	for index, total := 0, len(tokens); index < total; index++ {
		token := tokens[index]
		if isInBracket {
			if cur, ok := token.(*RightBracketToken); ok && cur.Stat.RBLevel == levels[0] && cur.Stat.RBSubLevel == levels[1] {
				stat := cur.Stat
				tokenNode, err := exp.toAst(subs, isInFunc)
				if err != nil {
					return nil, err
				}
				noRounds = append(noRounds, tokenNode)
				if isInFunc {
					noRounds = append(noRounds, bracketToOperator(")", stat))
				}
				isInBracket = false
				subs = nil
				isInFunc = false
			} else {
				subs = append(subs, token)
			}
		} else if cur, ok := token.(*LeftBracketToken); ok {
			stat := cur.Stat
			logics := stat.Logics
			if logics["IsFunc"] {
				noRounds = append(noRounds, bracketToOperator("(", stat))
				isInFunc = true
			} else {
				isInFunc = false
			}
			// for optimize, right tokens
			isNeedParseMore := true
			tokenLen := len(tokens)
			if rb, ok := tokens[index+1].(*RightBracketToken); ok {
				noRounds = append(noRounds, bracketToOperator(")", rb.Stat))
				index++
				isNeedParseMore = false
				isInBracket = false
				isInFunc = false
			} else if tokenLen > index+2 {
				if rb, ok := tokens[index+2].(*RightBracketToken); ok {
					noRounds = append(noRounds, tokens[index+1], bracketToOperator(")", rb.Stat))
					index += 2
					isNeedParseMore = false
					isInBracket = false
					isInFunc = false
				}
			}
			if isNeedParseMore {
				isInBracket = true
				levels = [2]int{stat.RBLevel, stat.RBSubLevel}
			}
		} else {
			noRounds = append(noRounds, token)
		}
	}
	lasts := []AnyToken{}
	isInBracket = false
	levels = [2]int{}
	// parse square brackets
	for index, total := 0, len(noRounds); index < total; index++ {
		token := noRounds[index]
		if isInBracket {
			if cur, ok := token.(*RightSquareBracketToken); ok && cur.Stat.SBLevel == levels[0] && cur.Stat.SBSubLevel == levels[1] {
				toAst, err := exp.toAst(subs, false)
				if err != nil {
					return nil, err
				}
				lasts = append(lasts, toAst)
				isInBracket = false
				subs = nil
				lasts = append(lasts, bracketToOperator("]", cur.Stat))
			} else {
				subs = append(subs, token)
			}
		} else if cur, ok := token.(*LeftSquareBracketToken); ok {
			stat := cur.Stat
			lasts = append(lasts, bracketToOperator("[", stat))
			// for optimize, right tokens
			if rsb, ok := noRounds[index+2].(*RightSquareBracketToken); ok {
				lasts = append(lasts, noRounds[index+1], bracketToOperator("]", rsb.Stat))
				index += 2
			} else {
				isInBracket = true
				levels = [2]int{stat.SBLevel, stat.SBSubLevel}
			}
		} else {
			lasts = append(lasts, token)
		}
	}
	// parse
	ops := []*OperatorToken{}
	parsed := []AnyToken{}
	total := len(lasts)
	for i := 0; i < total; i++ {
		token := lasts[i]
		if op, ok := token.(*OperatorToken); ok {
			var next AnyToken
			var nextNode *Node
			isNextTokenNode := false
			if i+1 < total {
				next = lasts[i+1]
				if cur, ok := next.(*TokenNode); ok {
					nextNode = cur.Node
					isNextTokenNode = true
				} else {
					nextNode = &Node{
						Type:  "raw",
						Token: next,
					}
				}
			}
			name := op.Name
			// parse unary
			if name == "!" {
				i++
				nextNode.Operator = name
				parsed = append(parsed, &TokenNode{
					Node: nextNode,
				})
			} else {
				// parse object and functions
				lastIndex := len(parsed) - 1
				last := parsed[lastIndex]
				switch name {
				case ".", "[":
					var node *Node
					if cur, ok := last.(*TokenNode); ok && cur.Type == "object" {
						node = cur.Node
					} else {
						var root *Node
						if ok {
							root = cur.Node
						} else {
							root = &Node{
								Type:  "raw",
								Token: last,
							}
						}
						node = &Node{
							Root: root,
							Type: "object",
						}
						if root.Operator != "" {
							node.Operator = root.Operator
							root.Operator = ""
						}
						curToken := &TokenNode{
							Type: "object",
							Node: node,
						}
						parsed[lastIndex] = curToken
					}
					nextNode.Operator = name
					node.Arguments = append(node.Arguments, nextNode)
					i++
				case "(":
					var root *Node
					if cur, ok := last.(*TokenNode); ok {
						root = cur.Node
					} else {
						root = &Node{
							Type:  "raw",
							Token: last,
						}
					}
					fnNode := &Node{
						Type: "function",
						Root: root,
					}
					if root.Operator != "" {
						fnNode.Operator = root.Operator
						root.Operator = ""
					}
					if isNextTokenNode {
						fnNode.Arguments = nextNode.Arguments
					} else {
						if op, ok := nextNode.Token.(*OperatorToken); ok && op.Name == ")" {
							// do nothing
						} else {
							fnNode.Arguments = append(fnNode.Arguments, nextNode)
						}
					}
					parsed[lastIndex] = &TokenNode{
						Node: fnNode,
						Type: "function",
					}
					i++
				case ")", "]":
					// do nothing
				default:
					stat := op.Stat
					stat.ParseIndex = len(parsed)
					parsed = append(parsed, op)
					ops = append(ops, op)
				}
			}
		} else {
			parsed = append(parsed, token)
		}
	}
	if len(parsed) == 1 {
		curToken := parsed[0]
		if result, ok := curToken.(*TokenNode); ok {
			return result, nil
		}
		return nil, fmt.Errorf("unexpect token,%#v", curToken)
	}
	// sort operators
	opPower := "**"
	sort.SliceStable(ops, func(i, j int) bool {
		a, b := ops[i], ops[j]
		// power form right to left
		if a.Name == opPower && b.Name == opPower {
			return a.Stat.ParseIndex > b.Stat.ParseIndex
		}
		return a.Exact.Priority > b.Exact.Priority
	})
	result := &Node{}
	isArg := false
	opLen := len(ops)
	for i, op := range ops {
		name := op.Name
		stat := op.Stat
		index := stat.ParseIndex
		prevIndex := index - 1
		nextIndex := index + 1
		left := getParsedNode(prevIndex, parsed)
		right := getParsedNode(nextIndex, parsed)
		if name == "|" {
			result = &Node{
				Type: "function",
				Root: right,
			}
			result.Arguments = append(result.Arguments, left)
		} else if name == ":" {
			result.Arguments = append(result.Arguments, right)
		} else if name == "," {
			if isArg {
				result.Arguments = append(result.Arguments, right)
			} else {
				result = &Node{
					Type: "function",
				}
				result.Arguments = append(result.Arguments, left, right)
				isArg = true
			}
		} else {
			result = &Node{
				Left:     left,
				Right:    right,
				Operator: name,
			}
			for j := i + 1; j < opLen; j++ {
				curOp := ops[j]
				curStat := curOp.Stat
				curIndex := curStat.ParseIndex
				if curIndex > index {
					curStat.ParseIndex -= 2
				}
			}
			parsed = append(parsed[:prevIndex+1], parsed[nextIndex+1:]...)
			parsed[prevIndex] = &TokenNode{
				Token: lastToken,
				Node:  result,
			}
		}
	}
	if isInFunc && !isArg {
		fnNode := &Node{
			Type: "function",
		}
		fnNode.Arguments = append(fnNode.Arguments, result)
		return &TokenNode{
			Token: lastToken,
			Node:  fnNode,
		}, nil
	}
	return &TokenNode{
		Token: lastToken,
		Node:  result,
	}, nil
}

func (exp *Expression) tokenize(context string) (tokens []AnyToken, err error) {
	if context == "" {
		return
	}
	parser := exp.Parser
	if exp.Initial {
		// initial stats
		parser.Reset()
		parser.Reuse()
	} else {
		exp.Initial = true
	}
	rns := []rune(context)
	total := len(rns)
	for i := 0; i < total; i++ {
		s := rns[i]
		if err = parser.Add(s); err != nil {
			return nil, err
		}
	}
	if parser.RBLevel != 0 {
		return nil, fmt.Errorf("wrong round bracket")
	} else if parser.SBLevel != 0 {
		return nil, fmt.Errorf("wrong square bracket")
	}
	// check if token is not complete
	current := parser.Current
	if current != nil {
		switch token := current.(type) {
		case *StringToken:
			return nil, fmt.Errorf("unclosed string token:%s", string(token.Stat.Values))
		case *SpaceToken:
			token.IsComplete = true
			return parser.Tokens, nil
		default:
			// need complete and validate
		}
	}
	if err = parser.Add(Space); err != nil {
		return nil, err
	}
	// ignore test space token, because it is not complete
	lasts := parser.Tokens
	// spew.Dump(lasts[0])
	EOF := &EOFToken{}
	if _, err = EOF.Validate(lasts); err != nil {
		return nil, err
	}
	return lasts, nil
}

// Parse parse expression
func (exp *Expression) Parse(context string) (*Node, error) {
	var tokens []AnyToken
	var err error
	if tokens, err = exp.tokenize(context); err != nil {
		return nil, err
	}
	loop := 0
	lasts := []AnyToken{}
	// remove all space tokens
	for _, token := range tokens {
		if _, isSpace := token.(*SpaceToken); !isSpace {
			stat := token.GetStat()
			stat.Index = loop
			token.SetStat(stat)
			lasts = append(lasts, token)
			loop++
		}
	}
	// parse to ast
	var lastToken *TokenNode
	if lastToken, err = exp.toAst(lasts, false); err != nil {
		return nil, err
	}
	return lastToken.Node, nil
}
