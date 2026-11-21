package pybot

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
)

var (
	mathNumberRE            = regexp.MustCompile(`\d+(?:\.\d+)?`)
	mathWordRE              = regexp.MustCompile(`(?i)\b(plus|add|added to|sum of|minus|subtract|subtracted from|times|multiplied by|multiply by|divided by|divide by|over|and)\b`)
	differenceBetweenRE     = regexp.MustCompile(`(?i)\bdifference between\s+(\d+(?:\.\d+)?)\s+and\s+(\d+(?:\.\d+)?)\b`)
	mathOperatorRE          = regexp.MustCompile(`(?i)\b(plus|added to|minus|subtracted from|times|multiplied by|multiply by|divided by|divide by|over)\b`)
	mathCommandRE           = regexp.MustCompile(`(?i)\b(calculate|compute|evaluate|solve|sum of|difference between|add|subtract|multiply|divide|how much is)\b`)
	mathIncompleteRequestRE = regexp.MustCompile(`(?i)\b(?:calculate|compute|evaluate|solve)\s+(?:this|that|it|something|the result|the answer)\b`)
	mathRootRE              = regexp.MustCompile(`(?i)\b(?:square root|sqrt)\b|√`)
	squareRootPhraseRE      = regexp.MustCompile(`(?i)\b(?:the\s+)?square root of\s+(-?\d+(?:\.\d+)?)`)
	sqrtSpacedRE            = regexp.MustCompile(`(?i)\bsqrt\s+(-?\d+(?:\.\d+)?)`)
	squareRootSymbolRE      = regexp.MustCompile(`√\s*(-?\d+(?:\.\d+)?)`)
	subtractFromRE          = regexp.MustCompile(`(?i)\b(?:subtract\s+(\d+(?:\.\d+)?)\s+from\s+(\d+(?:\.\d+)?)|(\d+(?:\.\d+)?)\s+subtracted from\s+(\d+(?:\.\d+)?))\b`)
	divideByRE              = regexp.MustCompile(`(?i)\bdivide\s+(\d+(?:\.\d+)?)\s+by\s+(\d+(?:\.\d+)?)\b`)
	multiplyByRE            = regexp.MustCompile(`(?i)\bmultiply\s+(\d+(?:\.\d+)?)\s+by\s+(\d+(?:\.\d+)?)\b`)
	rawMathExprRE           = regexp.MustCompile(`\d\s*[+*/-]\s*\d`)
	errMalformedExpression  = errors.New("malformed arithmetic expression")
	errDivisionByZero       = errors.New("division by zero")
)

var numberWords = map[string]string{
	"zero": "0", "one": "1", "two": "2", "three": "3", "four": "4", "five": "5",
	"six": "6", "seven": "7", "eight": "8", "nine": "9", "ten": "10", "eleven": "11",
	"twelve": "12", "thirteen": "13", "fourteen": "14", "fifteen": "15", "sixteen": "16",
	"seventeen": "17", "eighteen": "18", "nineteen": "19", "twenty": "20", "thirty": "30",
	"forty": "40", "fifty": "50", "sixty": "60", "seventy": "70", "eighty": "80", "ninety": "90",
}

var errNegativeSquareRoot = errors.New("square root of a negative number")

func solveArithmetic(text string) (Response, bool) {
	numberNormalized := replaceSpokenNumbers(text)
	numberCount := len(mathNumberRE.FindAllString(numberNormalized, -1))
	wordOperator := mathOperatorRE.MatchString(numberNormalized) && numberCount >= 2
	commandOperator := mathCommandRE.MatchString(text) && numberCount >= 2
	rawOperator := rawMathExprRE.MatchString(numberNormalized)
	rootOperator := mathRootRE.MatchString(text) && numberCount >= 1
	incompleteRequest := mathIncompleteRequestRE.MatchString(text)
	if !wordOperator && !commandOperator && !rawOperator && !rootOperator && !incompleteRequest {
		return Response{}, false
	}
	if incompleteRequest && !wordOperator && !rawOperator && !rootOperator {
		return Response{Text: "Give me an arithmetic expression, such as 12 plus 50.", Module: "reasoning/arithmetic", NeedsClarification: true}, true
	}
	normalized := normalizeMathOperators(numberNormalized)

	start := mathNumberRE.FindStringIndex(normalized)
	if start == nil {
		return Response{Text: "Give me an arithmetic expression, such as 12 plus 50.", Module: "reasoning/arithmetic", NeedsClarification: true}, true
	}
	if len(normalized) > 512 {
		return Response{Text: "That expression is too long for the local calculator.", Module: "reasoning/arithmetic", NeedsClarification: true}, true
	}
	startAt := start[0]
	probe := startAt - 1
	for probe >= 0 && (normalized[probe] == ' ' || normalized[probe] == '\t') {
		probe--
	}
	rootStart := strings.LastIndex(strings.ToLower(normalized[:start[0]]), "sqrt(")
	rootArgumentPrefix := ""
	if rootStart >= 0 {
		rootArgumentPrefix = strings.TrimSpace(normalized[rootStart+len("sqrt(") : start[0]])
	}
	if rootStart >= 0 && (rootArgumentPrefix == "" || rootArgumentPrefix == "-") {
		startAt = rootStart
	} else if probe >= 0 && normalized[probe] == '(' {
		startAt = probe
	}
	if startAt > 0 && normalized[startAt-1] == '-' {
		prefix := strings.TrimSpace(normalized[:startAt-1])
		if prefix == "" || strings.HasSuffix(prefix, "(") || strings.HasSuffix(prefix, "is") || strings.HasSuffix(prefix, "calculate") || strings.HasSuffix(prefix, "compute") {
			startAt--
			probe = startAt - 1
			for probe >= 0 && (normalized[probe] == ' ' || normalized[probe] == '\t') {
				probe--
			}
			if probe >= 0 && normalized[probe] == '(' {
				startAt = probe
			}
		}
	}
	parser := arithmeticParser{input: normalized[startAt:]}
	value, err := parser.parseExpression()
	if err == errDivisionByZero {
		return Response{Text: "Division by zero is undefined.", Module: "reasoning/arithmetic", NeedsClarification: true}, true
	}
	if err == errNegativeSquareRoot {
		return Response{Text: "The square root of a negative number is not a real number.", Module: "reasoning/arithmetic", NeedsClarification: true}, true
	}
	if err != nil || parser.operations == 0 {
		return Response{Text: "I couldn't parse that arithmetic expression. Use numbers with +, -, *, /, parentheses, or square roots.", Module: "reasoning/arithmetic", NeedsClarification: true}, true
	}
	remaining := strings.TrimSpace(parser.input[parser.position:])
	remaining = strings.Trim(remaining, "?.!,; \t")
	if remaining != "" {
		return Response{Text: "I couldn't parse the whole expression. Use numbers with +, -, *, /, parentheses, or square roots.", Module: "reasoning/arithmetic", NeedsClarification: true}, true
	}
	formatted := formatRational(value)
	if parser.approximated {
		formatted = "≈ " + formatted
	}
	return Response{Text: formatted, Module: "reasoning/arithmetic"}, true
}

func replaceSpokenNumbers(text string) string {
	for word, number := range numberWords {
		text = replaceWholeWord(text, word, number)
	}
	return text
}

func normalizeMathOperators(text string) string {
	text = squareRootPhraseRE.ReplaceAllString(text, "sqrt($1)")
	text = sqrtSpacedRE.ReplaceAllString(text, "sqrt($1)")
	text = squareRootSymbolRE.ReplaceAllString(text, "sqrt($1)")
	text = differenceBetweenRE.ReplaceAllString(text, "$1 - $2")
	text = subtractFromRE.ReplaceAllStringFunc(text, func(match string) string {
		parts := subtractFromRE.FindStringSubmatch(match)
		if parts[1] != "" {
			return parts[2] + " - " + parts[1]
		}
		return parts[4] + " - " + parts[3]
	})
	text = divideByRE.ReplaceAllString(text, "$1 / $2")
	text = multiplyByRE.ReplaceAllString(text, "$1 * $2")
	text = mathWordRE.ReplaceAllStringFunc(text, func(word string) string {
		switch strings.ToLower(word) {
		case "plus", "add", "added to", "sum of":
			return " + "
		case "minus", "subtract", "subtracted from":
			return " - "
		case "times", "multiplied by", "multiply by":
			return " * "
		case "divided by", "divide by", "over":
			return " / "
		case "and":
			return " + "
		default:
			return word
		}
	})
	return text
}

func replaceWholeWord(text, word, replacement string) string {
	pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
	return pattern.ReplaceAllString(text, replacement)
}

type arithmeticParser struct {
	input        string
	position     int
	operations   int
	approximated bool
}

func (p *arithmeticParser) parseExpression() (*big.Rat, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		p.skipSpaces()
		operator := p.peek()
		if operator != '+' && operator != '-' {
			return left, nil
		}
		p.position++
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		p.operations++
		if operator == '+' {
			left.Add(left, right)
		} else {
			left.Sub(left, right)
		}
	}
}

func (p *arithmeticParser) parseTerm() (*big.Rat, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		p.skipSpaces()
		operator := p.peek()
		if operator != '*' && operator != '/' {
			return left, nil
		}
		p.position++
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		p.operations++
		if operator == '*' {
			left.Mul(left, right)
		} else {
			if right.Sign() == 0 {
				return nil, errDivisionByZero
			}
			left.Quo(left, right)
		}
	}
}

func (p *arithmeticParser) parseUnary() (*big.Rat, error) {
	p.skipSpaces()
	operator := p.peek()
	if operator == '+' || operator == '-' {
		p.position++
		value, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		if operator == '-' {
			value.Neg(value)
		}
		return value, nil
	}
	return p.parseFactor()
}

func (p *arithmeticParser) parseFactor() (*big.Rat, error) {
	p.skipSpaces()
	if strings.HasPrefix(strings.ToLower(p.input[p.position:]), "sqrt") {
		p.position += len("sqrt")
		p.skipSpaces()
		if p.peek() != '(' {
			return nil, errMalformedExpression
		}
		p.position++
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipSpaces()
		if p.peek() != ')' {
			return nil, errMalformedExpression
		}
		p.position++
		root, approximated, err := rationalSquareRoot(value)
		if err != nil {
			return nil, err
		}
		p.operations++
		p.approximated = p.approximated || approximated
		return root, nil
	}
	if p.peek() == '(' {
		p.position++
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipSpaces()
		if p.peek() != ')' {
			return nil, errMalformedExpression
		}
		p.position++
		return value, nil
	}
	start := p.position
	for p.position < len(p.input) {
		ch := p.input[p.position]
		if (ch >= '0' && ch <= '9') || ch == '.' {
			p.position++
			continue
		}
		break
	}
	if start == p.position || p.position-start > 64 {
		return nil, errMalformedExpression
	}
	value := new(big.Rat)
	if _, ok := value.SetString(p.input[start:p.position]); !ok {
		return nil, errMalformedExpression
	}
	return value, nil
}

func rationalSquareRoot(value *big.Rat) (*big.Rat, bool, error) {
	if value.Sign() < 0 {
		return nil, false, errNegativeSquareRoot
	}
	numeratorRoot := new(big.Int).Sqrt(value.Num())
	denominatorRoot := new(big.Int).Sqrt(value.Denom())
	if new(big.Int).Mul(new(big.Int).Set(numeratorRoot), numeratorRoot).Cmp(value.Num()) == 0 &&
		new(big.Int).Mul(new(big.Int).Set(denominatorRoot), denominatorRoot).Cmp(value.Denom()) == 0 {
		return new(big.Rat).SetFrac(numeratorRoot, denominatorRoot), false, nil
	}

	const precision = 256
	input := new(big.Float).SetPrec(precision).SetRat(value)
	root := new(big.Float).SetPrec(precision).Sqrt(input)
	decimal := root.Text('f', 12)
	approximation := new(big.Rat)
	if _, ok := approximation.SetString(decimal); !ok {
		return nil, false, errMalformedExpression
	}
	return approximation, true, nil
}

func (p *arithmeticParser) skipSpaces() {
	for p.position < len(p.input) && (p.input[p.position] == ' ' || p.input[p.position] == '\t') {
		p.position++
	}
}

func (p *arithmeticParser) peek() byte {
	if p.position >= len(p.input) {
		return 0
	}
	return p.input[p.position]
}

func formatRational(value *big.Rat) string {
	if value.IsInt() {
		return value.Num().String()
	}
	denominator := new(big.Int).Set(value.Denom())
	for new(big.Int).Mod(denominator, big.NewInt(2)).Sign() == 0 {
		denominator.Div(denominator, big.NewInt(2))
	}
	for new(big.Int).Mod(denominator, big.NewInt(5)).Sign() == 0 {
		denominator.Div(denominator, big.NewInt(5))
	}
	if denominator.Cmp(big.NewInt(1)) == 0 {
		scale := 0
		denom := value.Denom().Int64()
		for denom%2 == 0 {
			denom /= 2
			scale++
		}
		for denom%5 == 0 {
			denom /= 5
			scale++
		}
		if scale > 12 {
			scale = 12
		}
		return strings.TrimRight(strings.TrimRight(value.FloatString(scale), "0"), ".")
	}
	return value.RatString() + " (approximately " + value.FloatString(6) + ")"
}
