// Package kdump like fmt.Println but more pretty and beautiful print Go values.
package kdump

import (
	"bytes"
	"io"

	"github.com/gookit/goutil/dump"
)

type KDumper dump.Dumper

func (d KDumper) WithOutput(output io.Writer) KDumper {
	d.Output = output
	return d
}

func (d KDumper) WithNoType() KDumper {
	d.NoType = true
	return d
}

func (d KDumper) WithNoColor() KDumper {
	d.NoColor = true
	return d
}

func (d KDumper) WithIndentLen(indentLen int) KDumper {
	d.IndentLen = indentLen
	return d
}

func (d KDumper) WithIndentChar(char byte) KDumper {
	d.IndentChar = char
	return d
}

func (d KDumper) WithMaxDepth(depth int) KDumper {
	d.MaxDepth = depth
	return d
}

func (d KDumper) WithShowFlag(flag int) KDumper {
	d.ShowFlag = flag
	return d
}

func (d KDumper) WithCallerSkip(callerSkip int) KDumper {
	d.CallerSkip = callerSkip
	return d
}

func (d KDumper) WithColorTheme(theme dump.Theme) KDumper {
	d.ColorTheme = theme
	return d
}

// Dump vars
func (d KDumper) Dump(vs ...interface{}) {
	(*dump.Dumper)(&d).Dump(vs...)
}

// Print vars. alias of Dump()
func (d KDumper) Print(vs ...interface{}) {
	(*dump.Dumper)(&d).Print(vs...)
}

// Println vars. alias of Dump()
func (d KDumper) Println(vs ...interface{}) {
	(*dump.Dumper)(&d).Println(vs...)
}

// Fprint print vars to io.Writer
func (d KDumper) Fprint(w io.Writer, vs ...interface{}) {
	(*dump.Dumper)(&d).Fprint(w, vs...)
}

// Format like fmt.Println, but the output is clearer and more beautiful
func (d KDumper) Format(vs ...interface{}) string {
	w := &bytes.Buffer{}
	(*dump.Dumper)(&d).Fprint(w, vs...)
	return w.String()
}

// Custom method outside the original dump package
// FormatN like fmt.Println, but the output is clearer and no color
func (d KDumper) FormatN(vs ...interface{}) string {
	w := &bytes.Buffer{}
	New().WithNoColor().Fprint(w, vs...)
	return w.String()
}
