package gribble

import (
	"fmt"
	"strings"
	"text/scanner"
)

// error returns either a verbose or a simple error, depending upon the value
// of 'verbose' in the parser state.
func (p *parser) error(pos scanner.Position, msg string) error {
	if p.verbose {
		return p.verboseError(pos, msg)
	}
	return p.simpleError(pos, msg)
}

// simpleError takes a position and returns a string with the line and column
// number, along with the error message.
func (p *parser) simpleError(pos scanner.Position, msg string) error {
	return e("Line %d, Column %d: %s", pos.Line, pos.Column, msg)
}

// verboseError takes a position and returns a string with the line that 
// contains that position with a '^' symbol placed beneathe the column 
// indicated by that position.
func (p *parser) verboseError(pos scanner.Position, msg string) error {
	// Sanitize the data first so that we can split on '\n' reliably.
	sanitized := strings.TrimSpace(p.original)
	sanitized = strings.Replace(sanitized, "\t", " ", -1)
	sanitized = strings.Replace(sanitized, "\r\n", "\n", -1)
	sanitized = strings.Replace(sanitized, "\r", "\n", -1)
	lines := strings.Split(sanitized, "\n")
	linei, coli := pos.Line-1, pos.Column-1

	if linei < 0 || linei >= len(lines) {
		panic(fmt.Sprintf("Line number %d is out of bounds.", linei))
	}
	line := lines[linei]

	// coli can point to one character after the last in 'line' (for EOF).
	if coli < 0 || coli > len(line) {
		panic(fmt.Sprintf("Column number %d is out of bounds.", coli))
	}

	caret := make([]rune, len(line)+1)
	for i := 0; i <= len(line); i++ {
		if i == coli {
			caret[i] = '^'
		} else {
			caret[i] = ' '
		}
	}
	return e("Line %d, Column %d: %s\n%s\n%s\n",
		pos.Line, pos.Column, msg, line, string(caret))
}

// parseError can be called at any point during parsing and an error will be
// appeneded to the parser's error list.
//
// 'kind' is the string representation of the value that was expected (but
// obviously wasn't found).
func (p *parser) parseError(kind string) {
	var found string
	if p.tok < 0 {
		found = fmt.Sprintf("'%s' (%s)",
			p.tokText, scanner.TokenString(p.tok))
	} else {
		found = fmt.Sprintf("'%s'", p.tokText)
	}

	err := p.error(p.Position,
		fmt.Sprintf("Expected '%s' but found %s.", kind, found))
	p.errors = append(p.errors, err)
}
