package parser

import (
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/lexer"
)

// Parse parses an input stylesheet.
func Parse(source *lexer.Source) *ast.Stylesheet {
	p := newParser(source)
	p.parse()
	return p.ss
}

func newParser(source *lexer.Source) *parser {
	return &parser{
		lexer: lexer.NewLexer(source),
		ss:    &ast.Stylesheet{},
	}
}

type parser struct {
	lexer *lexer.Lexer
	ss    *ast.Stylesheet
}

func (p *parser) parse() {
	for p.lexer.Current != lexer.EOF {
		switch p.lexer.Current {
		case lexer.At:
			p.parseAtRule()

		case lexer.Semicolon:
			p.lexer.Next()

		case lexer.CDO, lexer.CDC:
			// From https://www.w3.org/TR/css-syntax-3/#parser-entry-points,
			// we'll always assume we're parsing from the top-level, so we can discard CDO/CDC.
			p.lexer.Next()

		case lexer.Comment:
			p.ss.Nodes = append(p.ss.Nodes, &ast.Comment{
				Loc:  p.lexer.Location(),
				Text: p.lexer.CurrentString,
			})
			p.lexer.Next()

		default:
			p.ss.Nodes = append(p.ss.Nodes, p.parseQualifiedRule(false))
		}

	}
}

func isImportantString(in string) bool {
	return len(in) == 9 &&
		(in[0] == 'i' || in[0] == 'I') &&
		(in[1] == 'm' || in[1] == 'M') &&
		(in[2] == 'p' || in[2] == 'P') &&
		(in[3] == 'o' || in[3] == 'O') &&
		(in[4] == 'r' || in[4] == 'R') &&
		(in[5] == 't' || in[5] == 'T') &&
		(in[6] == 'a' || in[6] == 'A') &&
		(in[7] == 'n' || in[7] == 'N') &&
		(in[8] == 't' || in[8] == 'T')
}

// parseQualifiedRule parses a rule. If isKeyframes is set, the parser will assume
// all preludes are keyframes percentage selectors. Otherwise, it will assume
// the preludes are selector lists.
func (p *parser) parseQualifiedRule(isKeyframes bool) *ast.QualifiedRule {
	r := &ast.QualifiedRule{
		Loc: p.lexer.Location(),
	}

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.LCurly:
			block := &ast.DeclarationBlock{
				Loc: p.lexer.Location(),
			}

			r.Block = block
			p.lexer.Next()

			for p.lexer.Current != lexer.RCurly {
				decl := &ast.Declaration{
					Loc:      p.lexer.Location(),
					Property: p.lexer.CurrentString,
				}
				p.lexer.Expect(lexer.Ident)
				p.lexer.Expect(lexer.Colon)
			values:
				for {
					switch p.lexer.Current {
					case lexer.EOF:
						p.lexer.Errorf("unexpected EOF")
					case lexer.Semicolon:
						if len(decl.Values) == 0 {
							p.lexer.Errorf("declaration must have a value")
						}
						p.lexer.Next()
						block.Declarations = append(block.Declarations, decl)

						break values
					case lexer.Delim:
						if p.lexer.CurrentString != "!" {
							p.lexer.Errorf("unexpected token: %s", p.lexer.CurrentString)
						}
						p.lexer.Next()

						if !isImportantString(p.lexer.CurrentString) {
							p.lexer.Errorf("expected !important, unexpected token: %s", p.lexer.CurrentString)
						}
						p.lexer.Next()
						decl.Important = true

					case lexer.Comma:
						decl.Values = append(decl.Values, &ast.Comma{Loc: p.lexer.Location()})
						p.lexer.Next()

					default:
						decl.Values = append(decl.Values, p.parseValue(false))
					}
				}
			}
			p.lexer.Next()
			return r

		default:
			if isKeyframes {
				r.Prelude = p.parseKeyframeSelectorList()
				continue
			}

			r.Prelude = p.parseSelectorList()
		}
	}
}

func (p *parser) parseKeyframeSelectorList() *ast.KeyframeSelectorList {
	l := &ast.KeyframeSelectorList{
		Loc: p.lexer.Location(),
	}

	for {
		if p.lexer.Current == lexer.EOF {
			p.lexer.Errorf("unexpected EOF")
		}

		switch p.lexer.Current {
		case lexer.Percentage:
			l.Selectors = append(l.Selectors, &ast.Percentage{
				Loc:   p.lexer.Location(),
				Value: p.lexer.CurrentNumeral,
			})

		case lexer.Ident:
			if p.lexer.CurrentString != "from" && p.lexer.CurrentString != "to" {
				p.lexer.Errorf("unexpected string: %s. keyframe selector can only be from, to, or a percentage", p.lexer.CurrentString)
			}
			l.Selectors = append(l.Selectors, &ast.Identifier{
				Loc:   p.lexer.Location(),
				Value: p.lexer.CurrentString,
			})

		default:
			p.lexer.Errorf("unexepected token: %s. keyframe selector can only be from, to, or a percentage", p.lexer.Current.String())
		}
		p.lexer.Next()

		if p.lexer.Current == lexer.Comma {
			p.lexer.Next()
			continue
		}

		break
	}

	return l
}

// parseValue parses a possible ast value at the current position. Callers
// can set allowMathOperators if the enclosing context allows math expressions.
// See: https://www.w3.org/TR/css-values-4/#math-function.
func (p *parser) parseValue(allowMathOperators bool) ast.Value {
	switch p.lexer.Current {
	case lexer.Dimension:
		defer p.lexer.Next()
		return &ast.Dimension{
			Loc: p.lexer.Location(),

			Unit:  p.lexer.CurrentString,
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Percentage:
		defer p.lexer.Next()
		return &ast.Percentage{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Number:
		defer p.lexer.Next()
		return &ast.Number{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Ident:
		defer p.lexer.Next()
		return &ast.Identifier{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentString,
		}

	case lexer.Hash:
		defer p.lexer.Next()
		return &ast.HexColor{
			Loc:  p.lexer.Location(),
			RGBA: p.lexer.CurrentString,
		}

	case lexer.String:
		defer p.lexer.Next()
		return &ast.String{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentString,
		}

	case lexer.Delim:
		switch p.lexer.CurrentString {
		case "*", "/", "+", "-":
			if !allowMathOperators {
				p.lexer.Errorf("math operations are only allowed within: calc, min, max, or clamp")
				return nil
			}
			defer p.lexer.Next()

			return &ast.MathOperator{
				Loc:      p.lexer.Location(),
				Operator: p.lexer.CurrentString,
			}

		default:
			p.lexer.Errorf("unexpected token: %s", p.lexer.CurrentString)
			return nil
		}

	case lexer.FunctionStart:
		fn := &ast.Function{
			Loc:  p.lexer.Location(),
			Name: p.lexer.CurrentString,
		}
		p.lexer.Next()

	arguments:
		for {
			switch p.lexer.Current {
			case lexer.RParen:
				p.lexer.Next()
				break arguments
			case lexer.Comma:
				fn.Arguments = append(fn.Arguments, &ast.Comma{
					Loc: p.lexer.Location(),
				})
				p.lexer.Next()
			default:
				fn.Arguments = append(fn.Arguments, p.parseValue(fn.IsMath()))
			}
		}

		return fn
	default:
		p.lexer.Errorf("unknown token: %s|%s|%s", p.lexer.Current, p.lexer.CurrentString, p.lexer.CurrentNumeral)
		return nil
	}
}

func (p *parser) parseAtRule() {
	switch p.lexer.CurrentString {
	case "import":
		p.parseImportAtRule()

	case "media":
		p.parseMediaAtRule()

	case "keyframes", "-webkit-keyframes":
		p.parseKeyframes()

	default:
		p.lexer.Errorf("unsupported at rule: %s", p.lexer.CurrentString)
	}
}

// parseImportAtRule parses an import at rule. It roughly implements
// https://www.w3.org/TR/css-cascade-4/#at-import.
func (p *parser) parseImportAtRule() {
	prelude := &ast.String{
		Loc: p.lexer.Location(),
	}

	imp := &ast.AtRule{
		Loc:      p.lexer.Location(),
		Name:     p.lexer.CurrentString,
		Preludes: []ast.AtPrelude{prelude},
	}
	p.lexer.Next()

	switch p.lexer.Current {
	case lexer.URL:
		prelude.Value = p.lexer.CurrentString
		p.lexer.Next()

	case lexer.FunctionStart:
		if p.lexer.CurrentString != "url" {
			p.lexer.Errorf("@import target must be a url or string")
		}
		p.lexer.Next()

		prelude.Value = p.lexer.CurrentString
		p.lexer.Expect(lexer.String)
		p.lexer.Expect(lexer.RParen)

	case lexer.String:
		prelude.Value = p.lexer.CurrentString
		p.lexer.Expect(lexer.String)

	default:
		p.lexer.Errorf("unexpected import specifier")
	}

	imp.Preludes = append(imp.Preludes, p.parseMediaQueryList())

	p.ss.Nodes = append(p.ss.Nodes, imp)
}

// parseKeyframes parses a keyframes at rule. It roughly implements
// https://www.w3.org/TR/css-animations-1/#keyframes
func (p *parser) parseKeyframes() {
	r := &ast.AtRule{
		Loc:  p.lexer.Location(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	switch p.lexer.Current {
	case lexer.String:
		r.Preludes = append(r.Preludes, &ast.String{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentString,
		})

	case lexer.Ident:
		r.Preludes = append(r.Preludes, &ast.Identifier{
			Loc:   p.lexer.Location(),
			Value: p.lexer.CurrentString,
		})

	default:
		p.lexer.Errorf("unexpected token %s, expected string or identifier for keyframes", p.lexer.Current.String())
	}
	p.lexer.Next()

	block := &ast.QualifiedRuleBlock{
		Loc: p.lexer.Location(),
	}
	r.Block = block
	p.lexer.Expect(lexer.LCurly)
	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.RCurly:
			p.ss.Nodes = append(p.ss.Nodes, r)
			p.lexer.Next()
			return

		default:
			block.Rules = append(block.Rules, p.parseQualifiedRule(true))
		}
	}
}

// parseMediaAtRule parses a media at rule. It roughly implements
// https://www.w3.org/TR/mediaqueries-4/#media.
func (p *parser) parseMediaAtRule() {
	r := &ast.AtRule{
		Loc:  p.lexer.Location(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	r.Preludes = []ast.AtPrelude{p.parseMediaQueryList()}

	block := &ast.QualifiedRuleBlock{
		Loc: p.lexer.Location(),
	}
	r.Block = block
	p.lexer.Expect(lexer.LCurly)
	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.RCurly:
			p.ss.Nodes = append(p.ss.Nodes, r)
			p.lexer.Next()
			return

		default:
			block.Rules = append(block.Rules, p.parseQualifiedRule(false))
		}
	}
}

func (p *parser) parseMediaQueryList() *ast.MediaQueryList {
	l := &ast.MediaQueryList{
		Loc: p.lexer.Location(),
	}

	for {
		if p.lexer.Current == lexer.EOF {
			p.lexer.Errorf("unexpected EOF")
		}

		l.Queries = append(l.Queries, p.parseMediaQuery())

		if p.lexer.Current == lexer.Comma {
			p.lexer.Next()
			continue
		}

		break
	}

	return l
}

func (p *parser) parseMediaQuery() *ast.MediaQuery {
	q := &ast.MediaQuery{
		Loc: p.lexer.Location(),
	}

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")
			return q

		case lexer.LParen:
			q.Parts = append(q.Parts, p.parseMediaFeature())

		case lexer.Ident:
			q.Parts = append(q.Parts, p.parseValue(false).(*ast.Identifier))

		default:
			return q
		}
	}
}

func (p *parser) parseMediaFeature() ast.MediaFeature {
	startLoc := p.lexer.Location()
	p.lexer.Expect(lexer.LParen)

	firstValue := p.parseValue(false)

	switch p.lexer.Current {
	case lexer.RParen:
		p.lexer.Next()
		ident, ok := firstValue.(*ast.Identifier)
		if !ok {
			// XXX: this location is wrong. also, can't figure out type since we lost the lexer value.
			p.lexer.Errorf("expected identifier in media feature with no value")
		}

		return &ast.MediaFeaturePlain{
			Loc:      startLoc,
			Property: ident,
		}

	case lexer.Colon:
		p.lexer.Next()
		ident, ok := firstValue.(*ast.Identifier)
		if !ok {
			// XXX: this location is wrong. also, can't figure out type since we lost the lexer value.
			p.lexer.Errorf("expected identifier in non-range media feature")
		}

		secondValue := p.parseValue(false)

		p.lexer.Expect(lexer.RParen)
		return &ast.MediaFeaturePlain{
			Loc:      startLoc,
			Property: ident,
			Value:    secondValue,
		}

	case lexer.Delim:
		r := &ast.MediaFeatureRange{
			Loc:       startLoc,
			LeftValue: firstValue,
		}
		r.Operator = p.parseMediaRangeOperator()

		secondValue := p.parseValue(false)

		maybeIdent, ok := secondValue.(*ast.Identifier)
		if !ok {
			// If the first value was an identifier, then we'll call that the property.
			maybeIdent, ok := firstValue.(*ast.Identifier)
			if !ok {
				p.lexer.Errorf("expected identifier")
			}

			r.LeftValue = nil
			r.Property = maybeIdent
			r.RightValue = secondValue

			p.lexer.Expect(lexer.RParen)
			return r
		}
		r.Property = maybeIdent

		if p.lexer.Current == lexer.Delim {
			op := p.parseMediaRangeOperator()
			if op != r.Operator {
				p.lexer.Errorf("operators in a range must be the same")
			}
			r.RightValue = p.parseValue(false)
		}

		p.lexer.Expect(lexer.RParen)
		return r
	}

	p.lexer.Errorf("unexpected token: %s", p.lexer.Current.String())
	return nil
}

var (
	mediaOperatorLT  = "<"
	mediaOperatorLTE = "<="
	mediaOperatorGT  = ">"
	mediaOperatorGTE = ">="
)

func (p *parser) parseMediaRangeOperator() string {
	operator := p.lexer.CurrentString
	p.lexer.Next()

	if p.lexer.Current == lexer.Delim {
		if p.lexer.CurrentString != "=" || (operator != "<" && operator != ">") {
			p.lexer.Errorf("unexpected token: %s", p.lexer.Current.String())
		}

		p.lexer.Next()

		switch operator {
		case "<":
			return mediaOperatorLTE
		case ">":
			return mediaOperatorGTE
		default:
			p.lexer.Errorf("unknown operator: %s", operator)
		}
	}

	switch operator {
	case "<":
		return mediaOperatorLT
	case ">":
		return mediaOperatorGT
	default:
		p.lexer.Errorf("unknown operator: %s", operator)
		return ""
	}
}
