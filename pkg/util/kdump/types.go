// Package kdump like fmt.Println but more pretty and beautiful print Go values.
package kdump

import (
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
