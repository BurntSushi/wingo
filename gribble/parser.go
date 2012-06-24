package main

import (
	"fmt"
	"io"
	"strings"
	"text/scanner"
)

type parser struct {
	*scanner.Scanner
	tok rune
	tokText string
	errors []error
}

type command struct {
	name string
	params []param
}

func (cmd *command) String() string {
	params := make([]string, len(cmd.params))
	for i, param := range cmd.params {
		params[i] = fmt.Sprintf("%s", param)
	}
	return fmt.Sprintf("(%s %s)", cmd.name, strings.Join(params, " "))
}

// param is either a string, int, float64 or a command.
type param interface {}

func parse(reader io.Reader) (*command, error) {
	p := newParser(reader)
	cmd := p.command()
	if len(p.errors) == 0 && p.tok != scanner.EOF {
		p.error("EOF")
	}
	if len(p.errors) > 0 {
		reterrs := make([]string, len(p.errors))
		for i, err := range p.errors {
			reterrs[i] = err.Error()
		}
		return nil, fmt.Errorf(strings.Join(reterrs, "\n"))
	}
	return cmd, nil
}

func newParser(reader io.Reader) *parser {
	p := &parser{
		Scanner: &scanner.Scanner{},
	}
	p.Init(reader)
	return p
}

func (p *parser) Scan() rune {
	r := p.Scanner.Scan()
	p.tok, p.tokText = r, p.TokenText()
	return r
}

func (p *parser) error(kind string) {
	var found string
	if p.tok < 0 {
		found = fmt.Sprintf("'%s' (%s)",
			p.tokText, scanner.TokenString(p.tok))
	} else {
		found = fmt.Sprintf("'%s'", p.tokText)
	}
	p.errors = append(p.errors,
		fmt.Errorf("Line %d, Column %d: Expected '%s' but found %s.",
			p.Position.Line, p.Position.Column, kind, found))
}

func (p *parser) demands(c rune) {
	if p.tok != c {
		p.error(string([]rune{c}))
	}
	p.Scan()
}

func (p *parser) command() *command {
	switch tok := p.Scan(); tok {
	case '(':
		cmd := p.command()
		p.demands(')')
		return cmd
	case scanner.Ident:
		cmd := &command{
			name: p.TokenText(),
			params: p.params(),
		}
		return cmd
	}
	p.error("command")
	return nil
}

func (p *parser) params() []param {
	params := make([]param, 0, 4)
	tok := p.Scan()
	for tok != scanner.EOF {
		switch tok {
		case '(':
			params = append(params, p.command())
		case scanner.String:
			params = append(params, p.TokenText())
		case scanner.Int:
			params = append(params, p.TokenText())
		case scanner.Float:
			params = append(params, p.TokenText())
		default:
			return params
		}
		tok = p.Scan()
	}
	return params
}

func main() {
	// cmd := `(Workspace '\p50' (Move 5.0))` 
	cmd := `(Workspace (Move 5.0))`
	c, err := parse(strings.NewReader(cmd))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(c)
}

