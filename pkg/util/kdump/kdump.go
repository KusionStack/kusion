// Package kdump like fmt.Println but more pretty and beautiful print Go values.
package kdump

import (
	"bytes"
	"io"
	"os"

	"github.com/gookit/goutil/dump"
)

var std = (KDumper)(*dump.Std())

func New() KDumper {
	return (KDumper)(*dump.NewDumper(os.Stdout, 3))
}

// V like fmt.Println, but the output is clearer and more beautiful
func V(vs ...interface{}) {
	(*dump.Dumper)(&std).Dump(vs...)
}

// P like fmt.Println, but the output is clearer and more beautiful
func P(vs ...interface{}) {
	(*dump.Dumper)(&std).Print(vs...)
}

// Print like fmt.Println, but the output is clearer and more beautiful
func Print(vs ...interface{}) {
	(*dump.Dumper)(&std).Print(vs...)
}

// Println like fmt.Println, but the output is clearer and more beautiful
func Println(vs ...interface{}) {
	(*dump.Dumper)(&std).Println(vs...)
}

// Fprint like fmt.Println, but the output is clearer and more beautiful
func Fprint(w io.Writer, vs ...interface{}) {
	(*dump.Dumper)(&std).Fprint(w, vs...)
}

// Format like fmt.Println, but the output is clearer and more beautiful
func Format(vs ...interface{}) string {
	w := &bytes.Buffer{}
	(*dump.Dumper)(&std).Fprint(w, vs...)

	return w.String()
}

// Custom method outside the original dump package
// FormatN like fmt.Println, but the output is clearer and no color
func FormatN(vs ...interface{}) string {
	w := &bytes.Buffer{}
	d := (dump.Dumper)(New().WithNoColor())
	d.Fprint(w, vs...)

	return w.String()
}
