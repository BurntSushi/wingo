package gribble

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

// parser maintains the state of the Gribble parser. It embeds a Scanner value
// and keeps track of the current token and token text (since Peek/Next always
// look at the *next* token, which isn't what we want typically when reporting
// errors about the current token).
//
// errors is a slice of all errors encountered during parsing. Parsing never
// stops because of an error.
type parser struct {
	*scanner.Scanner
	original string
	verbose  bool
	tok      rune
	tokText  string
	errors   []error
}

// command represents a Gribble command invocation; which is simply a name and 
// a list of arguments.
type command struct {
	name   string
	params []Value
}

// String returns a recursively formatted command, Lisp style.
func (cmd *command) String() string {
	if len(cmd.params) == 0 {
		return fmt.Sprintf("(%s)", cmd.name)
	}

	params := make([]string, len(cmd.params))
	for i, param := range cmd.params {
		switch concrete := param.(type) {
		case int:
			params[i] = fmt.Sprintf("%d", concrete)
		case float64:
			params[i] = fmt.Sprintf("%0.2f", concrete)
		case string:
			params[i] = concrete
		case *command:
			params[i] = concrete.String()
		default:
			panic(fmt.Sprintf("Unexpected type: %T", concrete))
		}
	}
	return fmt.Sprintf("(%s %s)", cmd.name, strings.Join(params, " "))
}

// parse takes a command invocation string and returns its a command value
// representing it. Note that even when an error is returned, the command
// value is also returned since parsing doesn't stop on an error.
func parse(invocation string, verbose bool) (*command, error) {
	if len(strings.TrimSpace(invocation)) == 0 {
		return &command{}, fmt.Errorf("Empty strings are not valid commands.")
	}

	p := newParser(invocation, verbose)
	cmd := p.command()
	if len(p.errors) == 0 && p.tok != scanner.EOF {
		p.parseError("EOF")
	}
	if len(p.errors) > 0 {
		reterrs := make([]string, len(p.errors))
		for i, err := range p.errors {
			reterrs[i] = err.Error()
		}
		return cmd, e(strings.Join(reterrs, "\n"))
	}
	return cmd, nil
}

// newParser is a parser constructor and also initializes the Scanner.
func newParser(invocation string, verbose bool) *parser {
	p := &parser{
		Scanner:  &scanner.Scanner{},
		original: invocation,
		verbose:  verbose,
		errors:   make([]error, 0),
	}
	p.Init(strings.NewReader(invocation))
	p.Error = func(s *scanner.Scanner, msg string) {
		p.errors = append(p.errors, p.error(s.Position, msg))
	}
	return p
}

// Scan intercepts all Scan requests to update tok and tokText state.
func (p *parser) Scan() rune {
	r := p.Scanner.Scan()
	p.tok, p.tokText = r, p.TokenText()
	return r
}

// demands checks the current token for equality with 'c'. If they are equal,
// the scanner progresses unhindered. Otherwise, an error is logged and the
// scanner still progressed.
func (p *parser) demands(c rune) {
	if p.tok != c {
		p.parseError(string([]rune{c}))
	}
	p.Scan()
}

// command parses a command invocation. It can handle arbitrarily nested
// parantheses. An error is logged when something other than a scanner.Ident
// or a '(' is found.
func (p *parser) command() *command {
	switch tok := p.Scan(); tok {
	case '(':
		cmd := p.command()
		p.demands(')')
		return cmd
	case scanner.Ident:
		cmd := &command{
			name:   p.TokenText(),
			params: p.params(),
		}
		return cmd
	}
	p.parseError("command")
	return &command{}
}

// params parses a list of parameters to a command invocation. A parameter
// can be either a string, an integer, a float or a command (surrounded by
// parantheses). If a ')' or an EOF is found, parameter parsing stops.
// If anything else is found, an error is reported.
func (p *parser) params() []Value {
	params := make([]Value, 0, 4)
	tok := p.Scan()
	for tok != scanner.EOF {
		switch tok {
		case '(':
			params = append(params, p.command())
			p.demands(')')
			tok = p.tok
			continue
		case scanner.String:
			params = append(params, p.TokenText())
		case scanner.Int:
			if n, err := strconv.ParseInt(p.TokenText(), 0, 32); err != nil {
				p.parseError("integer")
				params = append(params, nil)
			} else {
				params = append(params, int(n))
			}
		case scanner.Float:
			if n, err := strconv.ParseFloat(p.TokenText(), 64); err != nil {
				p.parseError("float")
				params = append(params, nil)
			} else {
				params = append(params, float64(n))
			}
		case ')':
			return params
		default:
			p.parseError("parameter")
		}
		tok = p.Scan()
	}
	return params
}
