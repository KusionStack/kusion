// Package kdump like fmt.Println but more pretty and beautiful print Go values.
package kdump

import (
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
	std.Dump(vs...)
}

// P like fmt.Println, but the output is clearer and more beautiful
func P(vs ...interface{}) {
	std.Print(vs...)
}

// Print like fmt.Println, but the output is clearer and more beautiful
func Print(vs ...interface{}) {
	std.Print(vs...)
}

// Println like fmt.Println, but the output is clearer and more beautiful
func Println(vs ...interface{}) {
	std.Println(vs...)
}

// Fprint like fmt.Println, but the output is clearer and more beautiful
func Fprint(w io.Writer, vs ...interface{}) {
	std.Fprint(w, vs...)
}

// Format like fmt.Println, but the output is clearer and more beautiful
func Format(vs ...interface{}) string {
	return std.Format(vs...)
}

// Custom method outside the original dump package
// FormatN like fmt.Println, but the output is clearer and no color
func FormatN(vs ...interface{}) string {
	return std.FormatN(vs...)
}
