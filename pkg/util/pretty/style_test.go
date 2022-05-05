package pretty

import (
	"fmt"
	"testing"
)

func TestStyle(t *testing.T) {
	fmt.Println("Style:")
	defer fmt.Println("")

	fmt.Println(Cyan("This is a Cyan message"))
	fmt.Println(Gray("This is a Gray message"))
	fmt.Println(Blue("This is a Blue message"))
	fmt.Println(Black("This is a Black message"))
	fmt.Println(Green("This is a Green message"))
	fmt.Println(White("This is a White message"))
	fmt.Println(Yellow("This is a Yellow message"))
	fmt.Println(Magenta("This is a Magenta message"))
	fmt.Println(Normal("This is a Normal message"))
	fmt.Println(Red("This is a Red message"))
	fmt.Println(LightRed("This is a LightRed message"))
	fmt.Println(RedBold("This is a RedBold message"))
	fmt.Println(LightRedBold("This is a LightRedBold message"))
}
